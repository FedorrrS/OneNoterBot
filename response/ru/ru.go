package ru

const (
	GreetingRU = `
Привет, пользователь!
Пожалуйста, авторизуйся, потом можешь писать свои заметки 🗒🗒🗒`
	HelpRU = `
/start - Аутентификация, первое, что тебе нужно сделать
/notes - Все заметки будут показаны в виде списка
/clear - Удалить 1 заметку по номеру. После команды необходимо дать подтверждение и выбрать номер заметки, которая будет удалена,
	например: yes, 1
/clearall - Удалить все заметки
/whoami - Выводит имя пользователя, tg handle и tg link
	мы надеемся, что вы не забыли, кто вы 🙃
/help - Выводит эту информацию

Бот является очень очень MVP
По всем вопросам обращайтесь к @fselischev`
	AuthorizationFailedRU  = "Тебе нужно авторизоваться через команду /start"
	ClearVerificationRU    = `Точно хотите удалить выбранную заметку?`
	ClearallVerificationRU = `Точно хотите удалить все заметки?`
	CommandNotSupportedRU  = `
Это команда не поддерживается :( 
/help показывает все доступные команды`
	ClearYesRU          = "Заметка успешно удалена"
	ClearallYesRU       = "Ваши заметки успешно удалены!"
	ClearallNoRU        = "Окей, ваши заметки остались на своём месте"
	ClearallIncorrectRU = "Неправильный ответ.. Напишите yes или no"
	EmptyNotesRU        = "Пока что у тебя нет заметок :("
)