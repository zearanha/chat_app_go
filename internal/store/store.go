package store

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Message struct {
	ID        string    `bson:"_id,omitempty" json:"id,omitempty"`
	Username  string    `bson:"username" json:"username"`
	Room      string    `bson:"room" json:"room"`
	Content   string    `bson:"content" json:"content"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}

type User struct {
	ID           string `bson:"_id,omitempty" json:"id,omitempty"`
	Username     string `bson:"username" json:"username"`
	PasswordHash string `bson:"password_hash" json:"-"`
}

type Store struct {
	messages *mongo.Collection
	users    *mongo.Collection
}

func New(ctx context.Context, uri string) (*Store, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	db := client.Database("chatdb")
	messagesColl := db.Collection("messages")

	_, err = messagesColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "room", Value: 1}, {Key: "created_at", Value: -1}},
	})
	if err != nil {
		return nil, err
	}

	usersColl := db.Collection("users")
	_, err = usersColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return nil, err
	}

	return &Store{
		messages: messagesColl,
		users:    usersColl,
	}, nil
}

func (s *Store) CreateUser(ctx context.Context, username, passwordHash string) (User, error) {
	user := User{Username: username, PasswordHash: passwordHash}
	result, err := s.users.InsertOne(ctx, user)
	if err != nil {
		return user, err
	}
	if oid, ok := result.InsertedID.(interface{ Hex() string }); ok {
		user.ID = oid.Hex()
	}
	return user, nil
}

func (s *Store) FindUserByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	err := s.users.FindOne(ctx, bson.D{{Key: "username", Value: username}}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) SaveMessage(ctx context.Context, username, room, content string) (Message, error) {
	msg := Message{
		Username:  username,
		Room:      room,
		Content:   content,
		CreatedAt: time.Now(),
	}

	result, err := s.messages.InsertOne(ctx, msg)
	if err != nil {
		return msg, err
	}
	if oid, ok := result.InsertedID.(interface{ Hex() string }); ok {
		msg.ID = oid.Hex()
	}
	return msg, nil
}

func (s *Store) RecentMessages(ctx context.Context, room string, limit int64) ([]Message, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(limit)

	filter := bson.D{{Key: "room", Value: room}}

	cursor, err := s.messages.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (s *Store) ListRooms(ctx context.Context) ([]string, error) {
	rooms, err := s.messages.Distinct(ctx, "room", bson.D{})
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(rooms))
	for _, r := range rooms {
		if str, ok := r.(string); ok {
			result = append(result, str)
		}
	}
	return result, nil
}
