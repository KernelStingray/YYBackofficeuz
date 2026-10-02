package i18n

import (
	"testing"
)

func TestTranslator_AllEvents(t *testing.T) {
	tr := NewTranslator()

	tests := []struct {
		key      string
		lang     string
		expected string
	}{
		// booking_created
		{"booking_created_title", "uz", "Yangi bron"},
		{"booking_created_body", "uz", "Yo'lovchi safaringizga joy bron qildi"},
		{"booking_created_title", "ru", "Новая заявка"},
		{"booking_created_body", "ru", "Пассажир забронировал место на вашу поездку"},
		{"booking_created_title", "en", "New booking"},
		{"booking_created_body", "en", "A passenger booked a seat on your ride"},

		// booking_confirmed
		{"booking_confirmed_title", "uz", "Bron tasdiqlandi"},
		{"booking_confirmed_body", "uz", "Haydovchi bronigizni qabul qildi"},
		{"booking_confirmed_title", "ru", "Бронь подтверждена"},
		{"booking_confirmed_body", "ru", "Водитель принял вашу заявку"},

		// booking_rejected
		{"booking_rejected_title", "uz", "Bron rad etildi"},
		{"booking_rejected_body", "uz", "Haydovchi bronigizni qabul qilmadi. Boshqa safar tanlang"},
		{"booking_rejected_title", "ru", "Бронь отклонена"},

		// ride_cancelled
		{"ride_cancelled_title", "uz", "Safar bekor qilindi"},
		{"ride_cancelled_body", "uz", "Haydovchi safarni bekor qildi. Boshqa safar tanlang"},
		{"ride_cancelled_title", "ru", "Поездка отменена"},

		// booking_cancelled
		{"booking_cancelled_title", "uz", "Yo'lovchi bronni bekor qildi"},
		{"booking_cancelled_body", "uz", "O'rindiq bo'shadi"},
		{"booking_cancelled_title", "ru", "Пассажир отменил бронь"},

		// trial_started
		{"trial_started_title", "uz", "Sinov davri boshlandi"},
		{"trial_started_body", "uz", "7 kun davomida barcha imkoniyatlar ochiq"},
		{"trial_started_title", "ru", "Пробный период начался"},
		{"trial_started_body", "ru", "Все возможности открыты на 7 дней"},
		{"trial_started_title", "en", "Trial started"},
		{"trial_started_body", "en", "All features are open for 7 days"},

		// limit_exceeded
		{"limit_exceeded_title", "uz", "Limit tugadi"},
		{"limit_exceeded_body", "uz", "Safar paketingiz tugadi. Tarifni yangilang"},

		// Unknown language fallback to uz (default in Uzbekistan)
		{"booking_created_title", "unknown_lang", "Yangi bron"},
		{"ride_cancelled_title", "", "Safar bekor qilindi"},
	}

	for _, tt := range tests {
		res := tr.Translate(tt.lang, tt.key)
		if res != tt.expected {
			t.Errorf("Translate(%q, %q) = %q, expected %q", tt.lang, tt.key, res, tt.expected)
		}
	}
}
