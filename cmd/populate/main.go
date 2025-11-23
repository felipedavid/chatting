package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/felipedavid/chatting/storage"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Users         int
	Conversations int
	Messages      int
	BatchSize     int
	Workers       int
	DBURL         string
}

type PopulationStats struct {
	UsersGenerated         int
	ContactsGenerated      int
	ConversationsGenerated int
	MessagesGenerated      int
	StartTime              time.Time
	EndTime                time.Time
	mu                     sync.Mutex
}

func (s *PopulationStats) AddUsers(count int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.UsersGenerated += count
}

func (s *PopulationStats) AddContacts(count int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ContactsGenerated += count
}

func (s *PopulationStats) AddConversations(count int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ConversationsGenerated += count
}

func (s *PopulationStats) AddMessages(count int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.MessagesGenerated += count
}

func main() {
	// Parse command line flags
	config := parseFlags()

	ctx := context.Background()
	stats := &PopulationStats{StartTime: time.Now()}

	// Create connection pool for better performance
	pool, err := pgxpool.New(ctx, config.DBURL)
	if err != nil {
		log.Fatal("Failed to create connection pool:", err)
	}
	defer pool.Close()

	queries := storage.New(pool)
	gofakeit.Seed(time.Now().UnixNano())

	fmt.Printf("🚀 Starting data generation with config: %+v\n", config)

	// Generate users in batches
	users, err := generateUsersBatch(queries, ctx, config.Users, config.BatchSize, config.Workers, stats)
	if err != nil {
		log.Fatal("Failed to generate users:", err)
	}
	fmt.Printf("✅ Generated %d users\n", len(users))

	// Generate contacts with realistic social networks
	if err := generateContactsOptimized(queries, ctx, users, stats); err != nil {
		log.Fatal("Failed to generate contacts:", err)
	}
	fmt.Printf("✅ Generated %d contact relationships\n", stats.ContactsGenerated)

	// Generate conversations
	conversations, err := generateConversationsOptimized(queries, ctx, users, config.Conversations, stats)
	if err != nil {
		log.Fatal("Failed to generate conversations:", err)
	}
	fmt.Printf("✅ Generated %d conversations\n", len(conversations))

	// Generate messages in batches
	messageCount, err := generateMessagesBatch(queries, ctx, conversations, config.Messages, config.BatchSize, config.Workers, stats)
	if err != nil {
		log.Fatal("Failed to generate messages:", err)
	}
	fmt.Printf("✅ Generated %d messages\n", messageCount)

	stats.EndTime = time.Now()
	printFinalStats(stats)
}

func parseFlags() Config {
	var config Config
	flag.IntVar(&config.Users, "users", 10000, "Number of users to generate")
	flag.IntVar(&config.Conversations, "conversations", 100000, "Number of conversations to generate")
	flag.IntVar(&config.Messages, "messages", 3000000, "Number of messages to generate")
	flag.IntVar(&config.BatchSize, "batch-size", 1000, "Batch size for database operations")
	flag.IntVar(&config.Workers, "workers", 4, "Number of concurrent workers")
	flag.StringVar(&config.DBURL, "db-url", "postgres://postgres:postgres@127.0.0.1:5432/chatting", "Database connection URL")
	flag.Parse()
	return config
}

func printFinalStats(stats *PopulationStats) {
	duration := stats.EndTime.Sub(stats.StartTime)
	fmt.Printf("\n📊 Population Statistics:\n")
	fmt.Printf("⏱️  Total Duration: %v\n", duration)
	fmt.Printf("👥 Users Generated: %d\n", stats.UsersGenerated)
	fmt.Printf("📞 Contacts Generated: %d\n", stats.ContactsGenerated)
	fmt.Printf("💬 Conversations Generated: %d\n", stats.ConversationsGenerated)
	fmt.Printf("💭 Messages Generated: %d\n", stats.MessagesGenerated)

	if duration.Seconds() > 0 {
		fmt.Printf("⚡ Performance: %.2f records/second\n",
			float64(stats.UsersGenerated+stats.ContactsGenerated+stats.ConversationsGenerated+stats.MessagesGenerated)/duration.Seconds())
	}
}

// Batch user generation with concurrent workers
func generateUsersBatch(queries *storage.Queries, ctx context.Context, totalUsers, batchSize, workers int, stats *PopulationStats) ([]storage.User, error) {
	var allUsers []storage.User
	var mu sync.Mutex

	userBatches := make(chan int, workers)
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for batchSize := range userBatches {
				users, err := generateUserBatch(queries, ctx, batchSize)
				if err != nil {
					log.Printf("Error generating user batch: %v", err)
					continue
				}

				mu.Lock()
				allUsers = append(allUsers, users...)
				stats.AddUsers(len(users))
				mu.Unlock()

				if stats.UsersGenerated%1000 == 0 {
					fmt.Printf("👥 Generated %d users...\n", stats.UsersGenerated)
				}
			}
		}()
	}

	// Queue batches
	remaining := totalUsers
	for remaining > 0 {
		currentBatch := batchSize
		if remaining < batchSize {
			currentBatch = remaining
		}
		userBatches <- currentBatch
		remaining -= currentBatch
	}
	close(userBatches)

	wg.Wait()
	return allUsers, nil
}

func generateUserBatch(queries *storage.Queries, ctx context.Context, count int) ([]storage.User, error) {
	var users []storage.User
	professions := []string{"Software Engineer", "Teacher", "Photographer", "Student", "Doctor", "Nurse", "Lawyer", "Accountant", "Designer", "Manager", "Sales Rep", "Consultant", "Writer", "Artist", "Chef"}
	interests := []string{"technology", "travel", "photography", "cooking", "reading", "sports", "music", "movies", "gaming", "hiking", "art", "science", "politics", "fashion", "fitness"}

	for i := 0; i < count; i++ {
		profession := professions[rand.Intn(len(professions))]
		interest := interests[rand.Intn(len(interests))]

		// Generate realistic phone numbers from different countries
		var phone string
		country := rand.Intn(5)
		switch country {
		case 0: // US
			phone = fmt.Sprintf("+1%03d%03d%04d", rand.Intn(900)+100, rand.Intn(900)+100, rand.Intn(9000)+1000)
		case 1: // UK
			phone = fmt.Sprintf("+447%09d", rand.Intn(1000000000))
		case 2: // Germany
			phone = fmt.Sprintf("+4917%08d", rand.Intn(100000000))
		case 3: // France
			phone = fmt.Sprintf("+336%08d", rand.Intn(100000000))
		default: // Canada
			phone = fmt.Sprintf("+1%03d%03d%04d", rand.Intn(900)+100, rand.Intn(900)+100, rand.Intn(9000)+1000)
		}

		displayName := gofakeit.Name()
		about := fmt.Sprintf("%s passionate about %s", profession, interest)

		user, err := queries.CreateUser(ctx, storage.CreateUserParams{
			PhoneNumber: phone,
			DisplayName: pgtype.Text{String: displayName, Valid: true},
			About:       pgtype.Text{String: about, Valid: true},
		})
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

// Optimized contact generation with realistic social network patterns
func generateContactsOptimized(queries *storage.Queries, ctx context.Context, users []storage.User, stats *PopulationStats) error {
	// Social network patterns: people have different contact patterns
	// Some users are "social hubs" with many contacts, others are "lurkers" with few

	for i, user := range users {
		// Determine contact count based on user type
		var contactCount int
		userType := rand.Float32()

		if userType < 0.1 { // 10% are social hubs
			contactCount = rand.Intn(50) + 30 // 30-80 contacts
		} else if userType < 0.3 { // 20% are moderately social
			contactCount = rand.Intn(20) + 10 // 10-30 contacts
		} else { // 70% are regular users
			contactCount = rand.Intn(10) + 3 // 3-13 contacts
		}

		// Generate contacts with some preference for users with similar interests
		// (simplified: just random selection for now)
		usedContacts := make(map[pgtype.UUID]bool)
		usedContacts[user.ID] = true // Don't add self

		for j := 0; j < contactCount; j++ {
			// Find a suitable contact
			var contact storage.User
			attempts := 0
			for attempts < 10 {
				contactIdx := rand.Intn(len(users))
				contact = users[contactIdx]

				if !usedContacts[contact.ID] {
					break
				}
				attempts++
			}

			if attempts >= 10 {
				continue // Skip if we can't find a suitable contact
			}

			usedContacts[contact.ID] = true

			_, err := queries.AddContact(ctx, storage.AddContactParams{
				UserID:      user.ID,
				ContactID:   contact.ID,
				ContactName: pgtype.Text{String: contact.DisplayName.String, Valid: true},
			})
			if err != nil {
				continue // Skip duplicates
			}

			stats.AddContacts(1)
		}

		if i%1000 == 0 && i > 0 {
			fmt.Printf("📞 Generated contacts for %d users...\n", i)
		}
	}

	return nil
}

// Optimized conversation generation
func generateConversationsOptimized(queries *storage.Queries, ctx context.Context, users []storage.User, targetConversations int, stats *PopulationStats) ([]storage.Conversation, error) {
	var conversations []storage.Conversation
	var mu sync.Mutex

	// Generate 1-on-1 conversations first (majority)
	oneOnOneCount := int(float64(targetConversations) * 0.8)

	fmt.Printf("💬 Generating %d 1-on-1 conversations...\n", oneOnOneCount)

	// For each user, create conversations with their contacts
	conversationCount := 0
	for _, user := range users {
		if conversationCount >= oneOnOneCount {
			break
		}

		contacts, err := queries.ListUserContacts(ctx, user.ID)
		if err != nil {
			continue
		}

		// Create conversations with 60% of contacts
		for _, contact := range contacts {
			if conversationCount >= oneOnOneCount {
				break
			}

			if rand.Float32() < 0.6 { // 60% chance of conversation
				conv, err := queries.CreateConversation(ctx, storage.CreateConversationParams{
					IsGroup:   false,
					Title:     pgtype.Text{Valid: false},
					CreatedBy: pgtype.UUID{Valid: true, Bytes: user.ID.Bytes},
				})
				if err != nil {
					continue
				}

				// Add participants
				queries.AddConversationParticipant(ctx, storage.AddConversationParticipantParams{
					ConversationID: conv.ID,
					UserID:         user.ID,
					Role:           pgtype.Text{String: "member", Valid: true},
				})
				queries.AddConversationParticipant(ctx, storage.AddConversationParticipantParams{
					ConversationID: conv.ID,
					UserID:         contact.ContactID,
					Role:           pgtype.Text{String: "member", Valid: true},
				})

				mu.Lock()
				conversations = append(conversations, conv)
				stats.AddConversations(1)
				conversationCount++
				mu.Unlock()
			}
		}
	}

	// Generate group conversations
	groupCount := targetConversations - oneOnOneCount
	fmt.Printf("👥 Generating %d group conversations...\n", groupCount)

	groupNames := []string{
		"Family Chat", "Work Team", "College Friends", "Book Club", "Gaming Squad",
		"Travel Buddies", "Fitness Group", "Cooking Club", "Tech Talk", "Movie Night",
		"Study Group", "Project Team", "Neighborhood Watch", "Parent Group", "Hobby Club",
	}

	groupsCreated := 0
	for groupsCreated < groupCount {
		// Select random users for group
		groupSize := rand.Intn(8) + 3 // 3-11 members
		if groupSize > len(users) {
			groupSize = len(users)
		}

		shuffled := make([]storage.User, len(users))
		copy(shuffled, users)
		rand.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})

		groupMembers := shuffled[:groupSize]

		conv, err := queries.CreateConversation(ctx, storage.CreateConversationParams{
			IsGroup:   true,
			Title:     pgtype.Text{String: groupNames[rand.Intn(len(groupNames))], Valid: true},
			CreatedBy: pgtype.UUID{Valid: true, Bytes: groupMembers[0].ID.Bytes},
		})
		if err != nil {
			continue
		}

		// Add all members
		for _, member := range groupMembers {
			role := "member"
			if member.ID == conv.CreatedBy {
				role = "admin"
			}

			queries.AddConversationParticipant(ctx, storage.AddConversationParticipantParams{
				ConversationID: conv.ID,
				UserID:         member.ID,
				Role:           pgtype.Text{String: role, Valid: true},
			})
		}

		mu.Lock()
		conversations = append(conversations, conv)
		stats.AddConversations(1)
		groupsCreated++
		mu.Unlock()

		if groupsCreated%1000 == 0 {
			fmt.Printf("👥 Created %d group conversations...\n", groupsCreated)
		}
	}

	return conversations, nil
}

func generateUsers(queries *storage.Queries, ctx context.Context, count int) ([]storage.User, error) {
	var users []storage.User
	professions := []string{"Software Engineer", "Teacher", "Photographer", "Student", "Doctor"}
	interests := []string{"technology", "travel", "photography", "cooking", "reading"}

	for i := 0; i < count; i++ {
		profession := professions[rand.Intn(len(professions))]
		interest := interests[rand.Intn(len(interests))]

		phone := fmt.Sprintf("+1%03d%03d%04d",
			rand.Intn(900)+100, rand.Intn(900)+100, rand.Intn(9000)+1000)

		displayName := gofakeit.Name()
		about := fmt.Sprintf("%s passionate about %s", profession, interest)

		user, err := queries.CreateUser(ctx, storage.CreateUserParams{
			PhoneNumber: phone,
			DisplayName: pgtype.Text{String: displayName, Valid: true},
			About:       pgtype.Text{String: about, Valid: true},
		})
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

func generateContacts(queries *storage.Queries, ctx context.Context, users []storage.User) error {
	for i, user := range users {
		numContacts := rand.Intn(5) + 2 // 2-6 contacts

		for j := 0; j < numContacts; j++ {
			contactIdx := rand.Intn(len(users))
			if contactIdx == i {
				continue // Skip self
			}

			contact := users[contactIdx]

			_, err := queries.AddContact(ctx, storage.AddContactParams{
				UserID:      user.ID,
				ContactID:   contact.ID,
				ContactName: pgtype.Text{String: contact.DisplayName.String, Valid: true},
			})
			if err != nil {
				continue // Skip duplicates
			}
		}
	}

	return nil
}

func generateConversations(queries *storage.Queries, ctx context.Context, users []storage.User) ([]storage.Conversation, error) {
	var conversations []storage.Conversation

	// Generate 1-on-1 conversations
	for _, user := range users {
		contacts, err := queries.ListUserContacts(ctx, user.ID)
		if err != nil {
			return nil, err
		}

		for _, contact := range contacts {
			if rand.Float32() < 0.6 { // 60% chance of conversation
				conv, err := queries.CreateConversation(ctx, storage.CreateConversationParams{
					IsGroup:   false,
					Title:     pgtype.Text{Valid: false},
					CreatedBy: pgtype.UUID{Valid: true, Bytes: user.ID.Bytes},
				})
				if err != nil {
					continue
				}

				// Add participants
				queries.AddConversationParticipant(ctx, storage.AddConversationParticipantParams{
					ConversationID: conv.ID,
					UserID:         user.ID,
					Role:           pgtype.Text{String: "member", Valid: true},
				})
				queries.AddConversationParticipant(ctx, storage.AddConversationParticipantParams{
					ConversationID: conv.ID,
					UserID:         contact.ContactID,
					Role:           pgtype.Text{String: "member", Valid: true},
				})

				conversations = append(conversations, conv)
			}
		}
	}

	// Generate group conversations
	groupNames := []string{"Family Chat", "Work Team", "College Friends", "Book Club", "Gaming Squad"}
	numGroups := rand.Intn(3) + 2

	for i := 0; i < numGroups; i++ {
		// Select random users for group
		groupSize := rand.Intn(4) + 3
		groupMembers := make([]storage.User, 0, groupSize)

		shuffled := make([]storage.User, len(users))
		copy(shuffled, users)
		rand.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})

		groupMembers = shuffled[:groupSize]

		conv, err := queries.CreateConversation(ctx, storage.CreateConversationParams{
			IsGroup:   true,
			Title:     pgtype.Text{String: groupNames[rand.Intn(len(groupNames))], Valid: true},
			CreatedBy: pgtype.UUID{Valid: true, Bytes: groupMembers[0].ID.Bytes},
		})
		if err != nil {
			continue
		}

		// Add all members
		for _, member := range groupMembers {
			role := "member"
			if member.ID == conv.CreatedBy {
				role = "admin"
			}

			queries.AddConversationParticipant(ctx, storage.AddConversationParticipantParams{
				ConversationID: conv.ID,
				UserID:         member.ID,
				Role:           pgtype.Text{String: role, Valid: true},
			})
		}

		conversations = append(conversations, conv)
	}

	return conversations, nil
}

func generateMessages(queries *storage.Queries, ctx context.Context, conversations []storage.Conversation, users []storage.User) (int, error) {
	messageCount := 0

	for _, conv := range conversations {
		participants, err := queries.ListConversationParticipants(ctx, conv.ID)
		if err != nil {
			continue
		}

		numMessages := rand.Intn(20) + 10 // 10-30 messages

		for i := 0; i < numMessages; i++ {
			sender := participants[rand.Intn(len(participants))]

			messages := []string{
				"Hey, how are you doing?",
				"Did you see the news?",
				"I'm running a bit late, be there soon!",
				"Thanks for your help earlier!",
				"What's everyone up to this weekend?",
				"Check out this cool thing I found!",
				"Happy birthday! Hope you have an amazing day!",
				"Can't believe it's already been a week",
				"Just finished my project, it was awesome!",
				"Anyone free to chat?",
			}

			content := messages[rand.Intn(len(messages))]

			_, err := queries.CreateMessage(ctx, storage.CreateMessageParams{
				ConversationID: conv.ID,
				SenderID:       pgtype.UUID{Valid: true, Bytes: sender.UserID.Bytes},
				Content:        pgtype.Text{String: content, Valid: true},
				MessageType:    "text",
				ReplyToID:      pgtype.UUID{Valid: false},
			})
			if err != nil {
				continue
			}

			messageCount++
		}
	}

	return messageCount, nil
}

// Batch message generation with realistic content
func generateMessagesBatch(queries *storage.Queries, ctx context.Context, conversations []storage.Conversation, targetMessages int, batchSize int, workers int, stats *PopulationStats) (int, error) {
	messageBatches := make(chan []storage.Conversation, workers)
	var wg sync.WaitGroup
	var totalMessages int32

	// Start worker goroutines
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for convBatch := range messageBatches {
				for _, conv := range convBatch {
					// Generate messages for this conversation
					messagesGenerated, err := generateMessagesForConversation(queries, ctx, conv, stats)
					if err != nil {
						log.Printf("Error generating messages for conversation: %v", err)
						continue
					}

					atomic.AddInt32(&totalMessages, int32(messagesGenerated))

					if stats.MessagesGenerated%10000 == 0 {
						fmt.Printf("💭 Generated %d messages...\n", stats.MessagesGenerated)
					}
				}
			}
		}()
	}

	// Distribute conversations across workers
	batch := make([]storage.Conversation, 0, batchSize)
	messagesPerConv := targetMessages / len(conversations)

	for i, conv := range conversations {
		// Determine message count for this conversation based on conversation type and activity
		var messageCount int
		if conv.IsGroup {
			// Group conversations tend to have more messages
			messageCount = messagesPerConv + rand.Intn(messagesPerConv/2)
		} else {
			// 1-on-1 conversations
			messageCount = messagesPerConv/2 + rand.Intn(messagesPerConv/2)
		}

		// Ensure minimum messages per conversation
		if messageCount < 5 {
			messageCount = 5 + rand.Intn(15)
		}

		// Store message count in a temporary way (we'll generate them in the worker)
		// For now, just add to batch
		batch = append(batch, conv)

		if len(batch) >= batchSize || i == len(conversations)-1 {
			messageBatches <- batch
			batch = make([]storage.Conversation, 0, batchSize)
		}
	}

	close(messageBatches)
	wg.Wait()

	return int(totalMessages), nil
}

func generateMessagesForConversation(queries *storage.Queries, ctx context.Context, conv storage.Conversation, stats *PopulationStats) (int, error) {
	participants, err := queries.ListConversationParticipants(ctx, conv.ID)
	if err != nil {
		return 0, err
	}

	if len(participants) == 0 {
		return 0, nil
	}

	// Determine message count based on conversation type and activity level
	var messageCount int
	if conv.IsGroup {
		messageCount = rand.Intn(50) + 20 // 20-70 messages for groups
	} else {
		messageCount = rand.Intn(30) + 10 // 10-40 messages for 1-on-1
	}

	// Generate realistic message timing patterns
	// baseTime := time.Now().AddDate(0, 0, -rand.Intn(30)) // Random time within last 30 days

	for i := 0; i < messageCount; i++ {
		sender := participants[rand.Intn(len(participants))]

		// Generate realistic message content
		content := generateRealisticMessageContent()

		// Create message with some time spacing
		// messageTime := baseTime.Add(time.Duration(i*rand.Intn(300)) * time.Second)

		_, err := queries.CreateMessage(ctx, storage.CreateMessageParams{
			ConversationID: conv.ID,
			SenderID:       pgtype.UUID{Valid: true, Bytes: sender.UserID.Bytes},
			Content:        pgtype.Text{String: content, Valid: true},
			MessageType:    "text",
			ReplyToID:      pgtype.UUID{Valid: false},
		})
		if err != nil {
			continue
		}

		stats.AddMessages(1)
	}

	return messageCount, nil
}

func generateRealisticMessageContent() string {
	// Realistic message templates with variations
	templates := []string{
		"Hey %s! How are you doing?",
		"Did you see %s?",
		"I'm running a bit late, be there %s!",
		"Thanks for your help with %s!",
		"What's everyone up to %s?",
		"Check out this %s I found!",
		"Happy %s! Hope you have an amazing day!",
		"Can't believe it's already been %s",
		"Just finished my %s, it was awesome!",
		"Anyone free to %s?",
		"Good morning! Ready for %s?",
		"That %s was incredible!",
		"Looking forward to %s",
		"Sorry I'm late to the %s",
		"Congratulations on %s!",
		"Let me know about %s",
		"Thinking about %s",
		"Excited for %s!",
		"Thanks again for %s",
		"How did %s go?",
	}

	// Fillers for templates
	fillers := []string{
		"today", "yesterday", "tomorrow", "this week", "this weekend",
		"the news", "that movie", "the game", "the meeting", "the party",
		"soon", "in a few minutes", "later", "after work", "tomorrow",
		"the project", "the presentation", "the homework", "the cooking", "the planning",
		"this weekend", "tonight", "tomorrow", "next week", "for lunch",
		"cool thing", "awesome article", "funny video", "interesting post", "great deal",
		"birthday", "anniversary", "graduation", "promotion", "holiday",
		"a week", "a month", "a year", "so long", "forever",
		"project", "assignment", "workout", "meal", "trip",
		"chat", "hang out", "meet up", "call", "video call",
		"the day", "the presentation", "the interview", "the exam", "the workout",
		"experience", "concert", "show", "event", "performance",
		"the weekend", "vacation", "trip", "event", "celebration",
		"conversation", "discussion", "meeting", "party", "gathering",
		"achievement", "success", "win", "accomplishment", "milestone",
		"the details", "the plan", "the schedule", "the update", "the information",
		"you", "the future", "life", "everything", "the weekend",
		"the opportunity", "the adventure", "the challenge", "the experience", "the journey",
		"everything", "your help", "the support", "your time", "the opportunity",
		"the interview", "the exam", "the date", "the trip", "the meeting",
	}

	// Simple messages (no template needed)
	simpleMessages := []string{
		"Hey! 👋",
		"Good morning! ☀️",
		"Good night! 🌙",
		"How are you?",
		"What's up?",
		"LOL 😂",
		"Thanks! 😊",
		"You're welcome!",
		"See you later!",
		"Take care!",
		"Awesome! 🎉",
		"Sounds good!",
		"I agree",
		"Perfect!",
		"Great idea!",
		"Looking forward to it",
		"Can't wait!",
		"So excited!",
		"That's amazing!",
		"Wow! 😮",
	}

	// Choose message type
	messageType := rand.Intn(3)

	switch messageType {
	case 0: // Template message
		template := templates[rand.Intn(len(templates))]
		filler := fillers[rand.Intn(len(fillers))]
		return fmt.Sprintf(template, filler)
	case 1: // Simple message
		return simpleMessages[rand.Intn(len(simpleMessages))]
	default: // Emoji or short response
		shortResponses := []string{
			"👍", "👎", "❤️", "😂", "😊", "😢", "😮", "🎉", "🔥", "💯",
			"ok", "yes", "no", "maybe", "sure", "yep", "nope", "definitely",
			"absolutely", "probably", "I think so", "I don't think so",
		}
		return shortResponses[rand.Intn(len(shortResponses))]
	}
}
