package models

import "github.com/jmoiron/sqlx"

type MessageBundle struct {
	Message      Message `json:"message" xml:"message" form:"message" validate:"required"`
	RecipientIds []int   `json:"recipient_ids" xml:"recipient_ids" form:"recipient_ids"`
}

func (m *MessageBundle) Insert(db *sqlx.DB) (int, error) {
	id, err := m.Message.Insert(db)
	if err != nil {
		return -1, err
	}

	for _, recipientId := range m.RecipientIds {
		_, err = db.Exec("INSERT INTO recipient VALUES ($1, $2)", id, recipientId)
		if err != nil {
			return id, err
		}
	}
	return id, nil
}

func (m *MessageBundle) getAllRecipientIds(db *sqlx.DB) error {
	rows, err := db.Query("SELECT recipient_id FROM recipient WHERE message_id=$1", m.Message.ID)
	if err != nil {
		return err
	}
	for rows.Next() {
		var recipientId int
		err = rows.Scan(&recipientId)
		if err != nil {
			return nil
		}
		m.RecipientIds = append(m.RecipientIds, recipientId)
	}
	return nil
}

func (m *MessageBundle) GetAll(db *sqlx.DB, withRecipient bool) ([]*MessageBundle, error) {
	var messageBundles []*MessageBundle
	rows, err := m.Message.getAllRows(db, withRecipient)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var messageBundle MessageBundle
		err = rows.StructScan(&messageBundle.Message)
		if err != nil {
			return nil, err
		}
		err = messageBundle.getAllRecipientIds(db)
		if err != nil {
			return nil, err
		}
		messageBundles = append(messageBundles, &messageBundle)
	}
	return messageBundles, nil
}

func (m *MessageBundle) Get(db *sqlx.DB, withRecipient bool) error {
	err := m.Message.Get(db, withRecipient)
	if err != nil {
		return err
	}
	err = m.getAllRecipientIds(db)
	return err
}

func (m *MessageBundle) Update(db *sqlx.DB) error {
	err := m.Message.Update(db)
	if err != nil {
		return err
	}

	_, err = db.Exec("DELETE FROM recipient WHERE message_id=$1", m.Message.ID)
	if err != nil {
		return err
	}

	for _, recipientId := range m.RecipientIds {
		_, err = db.Exec("INSERT INTO recipient VALUES ($1, $2)", m.Message.ID, recipientId)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MessageBundle) name() {

}
