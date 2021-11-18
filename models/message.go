package models

import (
	"gorm.io/gorm"
	"time"
)

type Message struct {
	ID         uint           `gorm:"primarykey"`
	CreatedAt  time.Time      `json:"-"`
	UpdatedAt  time.Time      `json:"-"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	OwnerID    uint
	Text       string `validate:"required"`
	Palindrome bool
	Recipients []User `gorm:"many2many:recipients;"`
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
