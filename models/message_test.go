package models

import "testing"

func TestMessage_CheckPalindromeOdd(t *testing.T) {
	message := Message{
		Text:       "test 1234321 tset",
		Palindrome: false,
	}

	message.Check()

	if !message.Palindrome {
		t.Errorf("massage is a palindrome")
	}
}

func TestMessage_CheckPalindromeEven(t *testing.T) {
	message := Message{
		Text:       "test 123321 tset",
		Palindrome: false,
	}

	message.Check()

	if !message.Palindrome {
		t.Errorf("massage is a palindrome")
	}
}
func TestMessage_CheckNotPalindrome(t *testing.T) {
	message := Message{
		Text:       "test 1234 test",
		Palindrome: true,
	}

	message.Check()

	if message.Palindrome {
		t.Errorf("massage is not a palindrome")
	}
}
