package models

import "github.com/jmoiron/sqlx"

type Message struct {
	ID         int    `db:"id"`
	OwnerID    int    `db:"owner_id"`
	Text       string `db:"text" validate:"required"`
	Palindrome bool   `db:"palindrome"`
}

func (m *Message) Insert(db *sqlx.DB) (int, error) {
	row := db.QueryRow("INSERT INTO message (owner_id, text, palindrome) VALUES ($1,$2,$3) RETURNING id", m.OwnerID, m.Text, m.Palindrome)
	var id int
	err := row.Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (m *Message) Delete(db *sqlx.DB) (int64, error) {
	result, err := db.Exec("DELETE FROM message WHERE id=$1 AND owner_id=$2", m.ID, m.OwnerID)
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return affected, err
}

func (m *Message) getAllRows(db *sqlx.DB, withRecipient bool) (*sqlx.Rows, error) {
	if withRecipient {
		return db.Queryx("SELECT DISTINCT id,owner_id,text,palindrome FROM message LEFT JOIN recipient r on message.id = r.message_id WHERE owner_id=$1 OR recipient_id=$2", m.OwnerID, m.OwnerID)
	}
	return db.Queryx("SELECT * FROM message WHERE owner_id=$1", m.OwnerID)
}
func (m *Message) GetAll(db *sqlx.DB) ([]*Message, error) {
	var messages []*Message
	rows, err := db.Queryx("SELECT * FROM message WHERE owner_id=$1", m.OwnerID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var message Message
		err = rows.StructScan(&message)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &message)
	}
	return messages, nil
}

func (m *Message) Get(db *sqlx.DB, withRecipient bool) error {
	if withRecipient {
		return db.Get(m, "SELECT DISTINCT id,owner_id,text,palindrome FROM message LEFT JOIN recipient r on message.id = r.message_id WHERE id=$1 AND (owner_id=$2 OR recipient_id=$3)", m.ID, m.OwnerID, m.OwnerID)
	}
	return db.Get(m, "SELECT * FROM message WHERE id=$1 AND owner_id=$2", m.ID, m.OwnerID)
}

func (m *Message) Update(db *sqlx.DB) error {
	_, err := db.Exec("UPDATE message SET owner_id=$1, text=$2, palindrome=$3 WHERE id=$4", m.OwnerID, m.Text, m.Palindrome, m.ID)
	return err
}

func isPalindrome(s string) bool {
	palindrome := true
	numC := len(s)
	for i := 0; i < numC/2; i++ {
		palindrome = palindrome && s[i] == s[numC-i-1]
	}
	return palindrome
}

func (m *Message) Check() {
	m.Palindrome = isPalindrome(m.Text)
}
