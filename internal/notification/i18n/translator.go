package i18n

import (
	"fmt"
	"strings"
)

type Translator struct {
	messages map[string]map[string]string
}

func NewTranslator() *Translator {
	return &Translator{
		messages: map[string]map[string]string{
			"booking_created_title": {
				"uz": "Yangi bron",
				"ru": "Новая заявка",
				"en": "New booking",
			},
			"booking_created_body": {
				"uz": "Yo'lovchi safaringizga joy bron qildi",
				"ru": "Пассажир забронировал место на вашу поездку",
				"en": "A passenger booked a seat on your ride",
			},
			"booking_confirmed_title": {
				"uz": "Bron tasdiqlandi",
				"ru": "Бронь подтверждена",
				"en": "Booking confirmed",
			},
			"booking_confirmed_body": {
				"uz": "Haydovchi bronigizni qabul qildi",
				"ru": "Водитель принял вашу заявку",
				"en": "The driver accepted your booking",
			},
			"booking_rejected_title": {
				"uz": "Bron rad etildi",
				"ru": "Бронь отклонена",
				"en": "Booking declined",
			},
			"booking_rejected_body": {
				"uz": "Haydovchi bronigizni qabul qilmadi. Boshqa safar tanlang",
				"ru": "Водитель не принял вашу заявку. Выберите другую поездку",
				"en": "The driver did not accept your booking. Please pick another ride",
			},
			"ride_cancelled_title": {
				"uz": "Safar bekor qilindi",
				"ru": "Поездка отменена",
				"en": "Ride cancelled",
			},
			"ride_cancelled_body": {
				"uz": "Haydovchi safarni bekor qildi. Boshqa safar tanlang",
				"ru": "Водитель отменил поездку. Выберите другую",
				"en": "The driver cancelled the ride. Please pick another one",
			},
			"booking_cancelled_title": {
				"uz": "Yo'lovchi bronni bekor qildi",
				"ru": "Пассажир отменил бронь",
				"en": "Passenger cancelled",
			},
			"booking_cancelled_body": {
				"uz": "O'rindiq bo'shadi",
				"ru": "Место освободилось",
				"en": "The seat is free again",
			},
			"trial_started_title": {
				"uz": "Sinov davri boshlandi",
				"ru": "Пробный период начался",
				"en": "Trial started",
			},
			"trial_started_body": {
				"uz": "7 kun davomida barcha imkoniyatlar ochiq",
				"ru": "Все возможности открыты на 7 дней",
				"en": "All features are open for 7 days",
			},
			"limit_exceeded_title": {
				"uz": "Limit tugadi",
				"ru": "Лимит исчерпан",
				"en": "Limit reached",
			},
			"limit_exceeded_body": {
				"uz": "Safar paketingiz tugadi. Tarifni yangilang",
				"ru": "Пакет поездок закончился. Обновите тариф",
				"en": "Your ride package is used up. Renew your plan",
			},
			"subscription_activated": {
				"ru": "Подписка %s успешно активирована!",
				"uz": "%s obunasi muvaffaqiyatli faollashtirildi!",
				"en": "Subscription %s successfully activated!",
			},
			"new_message": {
				"ru": "У вас новое сообщение!",
				"uz": "Sizda yangi xabar bor!",
				"en": "You have a new message!",
			},
			"ride_started": {
				"ru": "Ваша поездка началась.",
				"uz": "Sizning sayohatingiz boshlandi.",
				"en": "Your ride has started.",
			},
			"booking.confirmed": {
				"ru": "Ваше бронирование подтверждено.",
				"uz": "Sizning band qilingan joyingiz tasdiqlandi.",
				"en": "Your booking is confirmed.",
			},
			"ride.cancelled": {
				"ru": "Поездка была отменена водителем.",
				"uz": "Safarga haydovchi tomonidan bekor qilindi.",
				"en": "The ride was cancelled by the driver.",
			},
			"ride_cancelled": {
				"ru": "Поездка была отменена.",
				"uz": "Sayohat bekor qilindi.",
				"en": "The ride was cancelled.",
			},
			"ride_created": {
				"ru": "Ваша поездка успешно создана.",
				"uz": "Sayohat muvaffaqiyatli yaratildi.",
				"en": "Your ride was successfully created.",
			},
			"negotiation_accepted": {
				"ru": "Ваше предложение цены принято!",
				"uz": "Sizning narx taklifingiz qabul qilindi!",
				"en": "Your price negotiation was accepted!",
			},
			"payment.failed": {
				"ru": "Ошибка при списании средств.",
				"uz": "Mablag'larni yechishda xatolik yuz berdi.",
				"en": "Payment failed.",
			},
			"route_match": {
				"ru": "Найден водитель по вашему маршруту!",
				"uz": "Sizning yo'nalishingiz bo'yicha haydovchi topildi!",
				"en": "A driver was found for your route!",
			},
			"subscription_expired": {
				"ru": "Срок действия вашей подписки истек.",
				"uz": "Sizning obunangiz muddati tugadi.",
				"en": "Your subscription has expired.",
			},
			"car_approved": {
				"ru": "Ваш автомобиль %s успешно одобрен!",
				"uz": "Sizning %s avtomobilingiz muvaffaqiyatli tasdiqlandi!",
				"en": "Your car %s has been successfully approved!",
			},
			"car_rejected": {
				"ru": "К сожалению, ваш автомобиль %s был отклонен.",
				"uz": "Kechirasiz, sizning %s avtomobilingiz rad etildi.",
				"en": "Unfortunately, your car %s has been rejected.",
			},
			"passenger_no_show": {
				"ru": "Пассажир не явился. Ваша поездка обновлена.",
				"uz": "Yo'lovchi kelmadi. Sizning sayohatingiz yangilandi.",
				"en": "Passenger no-show. Your ride has been updated.",
			},
			"negotiation_requested": {
				"ru": "Новое предложение цены от пассажира!",
				"uz": "Yo'lovchidan yangi narx taklifi!",
				"en": "New price negotiation requested by passenger!",
			},
			"ride_completed": {
				"ru": "Поездка успешно завершена. Спасибо, что выбираете нас!",
				"uz": "Sayohat muvaffaqiyatli yakunlandi. Bizni tanlaganingiz uchun tashakkur!",
				"en": "Ride completed successfully. Thank you for choosing us!",
			},
		},
	}
}

// Translate returns the formatted message in the requested language.
// If the language is not supported, it falls back to "uz" (default in Uzbekistan).
func (t *Translator) Translate(lang, key string, args ...interface{}) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if strings.HasPrefix(lang, "uz") {
		lang = "uz"
	} else if strings.HasPrefix(lang, "ru") {
		lang = "ru"
	} else if strings.HasPrefix(lang, "en") {
		lang = "en"
	} else {
		lang = "uz" // Default in Uzbekistan
	}

	langMap, exists := t.messages[key]
	if !exists {
		return key // fallback to key itself if translation not found
	}

	msg, exists := langMap[lang]
	if !exists {
		msg, exists = langMap["uz"]
		if !exists {
			msg, exists = langMap["ru"]
			if !exists {
				return key
			}
		}
	}

	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}
	return msg
}
