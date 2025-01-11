package main

import (
	"kubete_torrentBot/botlogic"
	"kubete_torrentBot/remote"
	"log"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var msgs []tgbotapi.MessageConfig

var mutex sync.Mutex

func main() {
	// запускаем таймер в отдельной горутине
	go Timer()

	// запускаем таймер в отдельной горутине
	go Timer()

	// создаём бота по зарегестрированному API
	bot, err := tgbotapi.NewBotAPI("7913546799:AAFUknlXWVYhsqgcOzq8P3S3ZLItye3-G8g")
	if err != nil {
		log.Panic(err)
	}

	// режим отладки для логов
	bot.Debug = true

	// канал обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// список новых сообщений
	updates := bot.GetUpdatesChan(u)
	// обработка сообщений
	for {
		for update := range updates {
			// отсылаем сообщения по таймеру
			// блокируем чтобы msgs не изменился в процессе

			mutex.Lock()
			if len(msgs) > 0 {
				for n, msg := range msgs {
					bot.Send(msg)
					msgs = append(msgs[:n], msgs[n+1:]...)
				}
			}
			mutex.Unlock()

			// обрабатываем остальные сообщения
			if update.Message != nil {

				// вывод логов
				log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

				// основная часть
				var msg tgbotapi.MessageConfig

				chat_id := update.Message.Chat.ID
				status := botlogic.Get_status(chat_id)
				command := update.Message.Command()
				args := update.Message.CommandArguments()

				// не даём выполнять команды неизвестным и анонимным
				if status == "Неизвестный" || status == "Анонимный" {
					if command != "start" && command != "help" && command != "status" && command != "login" && command != "anonme" && command != "aume" {
						command = "login"
						msg1 := tgbotapi.NewMessage(update.Message.Chat.ID, "Сначала пройдите авторизацию")
						bot.Send(msg1)
					}
				}

				// коммады
				switch command {
				case "start":
					start_text :=
						`Это телеграм-бот для проведения массовых опросов и тестирований.
					Узнайте о коммандах введя /help.
					Не забудьте пройти авторизацию: /login`
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, start_text)
				case "help":
					head :=
						`Вот что умеет этот бот:
					переключайтесь между страницами (1-7):
					/help <номер страницы>` + "\n"

					help1 :=
						`1. Команды бота и авторизация:
					/start - начало работы
					/help - возможности бота
					/status - статус пользователя (Неизвестный/Анонимный/Авторизованный)
					/login - вход
					/login type - регистрация
					/logout - завершение сеанса
					/logout all=true - на всех устройствах`

					help2 :=
						`2. Команды, связанные с пользователями:
					(<id> - id пользователя)
					/userList - список пользователей
					/name <id> - посмотреть ФИО пользователя
					/nameChange <id> - изменить ФИО пользователя
					/userData <id> - информация о пользователе (курсы, оценки, тесты)
					/role <id> - посмотреть роль пользователя (студент/преподаватель/админ)
					/roleChange <id> <role> - изменить роль пользователя
					/blockInfo <id> - заблокирован ли пользователь
					/block <id> - заблокировать пользователя
					/unblock <id> - разблокировать пользователя`

					help3 :=
						`3. Комманды, связанные с дисциплинами:
					(<id> - id дисциплины)
					/courseList - список дисциплин
					/course <id> - полная иформация о курсе
					/courseChangeName <id> - изменить название дисциплины
					/courseChangeInfo <id> - изменить описание дисциплины
					/courseTestList <id> - список тестов дисциплины
					/testActive <id> - активен ли тест
					/testActivate <id> - активировать тест
					/testDeactivate <id> - деактивировать тест
					/testAdd <id> - добавить пустой тест
					/testDelete <id> - удалить тест
					/courseStudentList <id> - список студентов дисциплины
					/courseStudentAdd <id дисциплины> <id пользователя> - записать студента на дисциплину
					/courseStudentDelete <id дисциплины> <id пользователя> - отчислить студента с дисциплины
					/courseAdd - создать дисциплину
					/courseDelete <id> - удалить дисциплину`

					help4 :=
						`4. Команды, связанные с вопросами:
					(<id> - id вопроса)
					/questList <id пользователя> <id теста>- список всех ваших вопросов
					/questInfo <id вопроса> <id ответа> - информация о вопросе (Название, Текст, id, Ответ)
					/questUpdate <id вопроса> <id ответа> - изменить вопрос
					/questCreate <id теста> - создать вопрос
					/questDelete <id> - удалить вопрос`

					help5 :=
						`5. Команды, связанные с тестами:
					(<id> - id теста)
					/testQuestDelete <id теста> <id вопроса> - удалить вопрос из теста (только если ещё не было попыток прохождения)
					/testQuestAdd <id теста> <id вопроса> - добавить вопрос в тест
					/testProcedureChange <id> - изменить порядок вопросов в тесте
					/testAnswerUsersList <id> - список пользователей, прошедших тест
					/testUserAnswers <id пользователя> <id теста>, - посмотореть ответы пользователя на тест
					/testUserMark <id пользователя> <id теста> - посмотреть оценку за тест`

					help6 :=
						`6. Комманды, связанные с попытками
					(<id> - id теста)
					/attempCreate <id> - создать попытку
					/attempUpdate <id> - изменить попытку
					/attempEnd <id> - завершить попытку
					/attempRead <id пользователя> <id теста> - посмтореть попытку пользователя`

					help7 :=
						`7. Комманды, связанные с ответами
					/answerCreate <id теста> <id ответа> - создать ответ
					/answerRead <id теста> <id ответа> - посмотреть ответ
					/answerChange <id теста> <id ответа> - изменить ответ
					/answerDelete <id теста> <id ответа> - удалить ответ`

					help_text := head
					switch args {
					case "2":
						help_text += help2
					case "3":
						help_text += help3
					case "4":
						help_text += help4
					case "5":
						help_text += help5
					case "6":
						help_text += help6
					case "7":
						help_text += help7
					default:
						help_text += help1
					}

					msg = tgbotapi.NewMessage(update.Message.Chat.ID, help_text)
				case "status":
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Сейчас вы "+status+" пользователь")
				case "login":
					switch args {
					case "type":
						switch status {
						case "Авторизованный":
							msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Вы уже авторизованы!")
						default:
							msg = tgbotapi.NewMessage(update.Message.Chat.ID, botlogic.Login_type(chat_id))
						}
					default:
						switch status {
						case "Неизвестный":

							git_ref := remote.SendAu("/GET github")
							ya_ref := remote.SendAu("/GET yandex")
							if git_ref == "1" {
								git_ref = "https://your-api-endpoint.com/auth/github"
							}
							if ya_ref == "1" {
								ya_ref = "https://your-api-endpoint.com/auth/yandex"
							}
							keyboard := tgbotapi.NewInlineKeyboardMarkup(
								tgbotapi.NewInlineKeyboardRow(
									tgbotapi.NewInlineKeyboardButtonURL("🐱Github", git_ref),
								),
								tgbotapi.NewInlineKeyboardRow(
									tgbotapi.NewInlineKeyboardButtonURL("📕Яндекс ID", ya_ref),
								),
								tgbotapi.NewInlineKeyboardRow(
									tgbotapi.NewInlineKeyboardButtonData("🧾Код", "button_code"),
								),
							)
							msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Авторизуйтесь через:")
							msg.ReplyMarkup = keyboard
						case "Авторизованный":
							msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Вы уже авторизованы!")
						case "Анонимный":
							msg = tgbotapi.NewMessage(update.Message.Chat.ID, botlogic.Login(chat_id))
						}
					}
				case "logout":
					if args == "all=true" {
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, botlogic.Logout_all(chat_id))
					} else {
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, botlogic.Logout(chat_id))
					}
				case "userList":
					text := botlogic.SendToMain(chat_id, "users:list:read")
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "name":
					text := botlogic.SendToMain(chat_id, "users:fullname:read:other "+"id_user="+args)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "nameChange":
					text := botlogic.SendToMain(chat_id, "users:fullname:write:other "+"id_user="+args)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "userData":
					text := botlogic.SendToMain(chat_id, "users:data:read:other "+"id_user="+args)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "role":
					text := botlogic.SendToMain(chat_id, "users:roles:read:other "+"id_user="+args)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "roleChange":
					s := strings.SplitAfter(args, " ")
					if len(s) != 2 {
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Неверное число аргументов")
						break
					}
					text := botlogic.SendToMain(chat_id, "users:roles:write:other "+"id_user="+s[0]+" role="+s[1])
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "blockInfo":
					text := botlogic.SendToMain(chat_id, "users:block:read:other "+"id_user="+args)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "block":
					text := botlogic.SendToMain(chat_id, "users:block:write:other "+"id_user="+args)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "unblock":
					text := botlogic.SendToMain(chat_id, "users:unblock:write:other "+"id_user="+args)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "courseList":
					text := botlogic.SendToMain(chat_id, "course:list:read")
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "course":
					text := botlogic.SendToMain(chat_id, "course:data:read "+"id_course="+args)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "courseChangeName":
					text := botlogic.SendToMain(chat_id, "course:name:write "+"id_course="+args)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "courseChangeInfo":
					text := botlogic.SendToMain(chat_id, "course:info:write "+"id_course="+args)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "courseTestList":
					text := botlogic.SendToMain(chat_id, "course:testList:read:other "+"id_course="+args)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "testActive":
					text := botlogic.SendToMain(chat_id, "course:test:read:other "+"id_test="+args)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "testActivate":
					text := botlogic.SendToMain(chat_id, "course:test:activate "+"id_test="+args)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "testDeactivate":
					text := botlogic.SendToMain(chat_id, "course:test:deactivate "+"id_test="+args)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "testAdd":
					text := botlogic.SendToMain(chat_id, "course:test:add "+"id_course="+args)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "ID нового теста: "+text)
				case "testDelete":
					text := botlogic.SendToMain(chat_id, "course:test:del "+"id_course="+args)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "courseStudentList":
					text := botlogic.SendToMain(chat_id, "course:userList:read "+"id_course="+args)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "courseStudentAdd":
					s := strings.SplitAfter(args, " ")
					if len(s) != 2 {
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Неверное число аргументов")
						break
					}
					text := botlogic.SendToMain(chat_id, "course:user:add "+"id_course="+s[0]+" id_user="+s[1])
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "courseStudentDelete":
					s := strings.SplitAfter(args, " ")
					if len(s) != 2 {
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Неверное число аргументов")
						break
					}
					text := botlogic.SendToMain(chat_id, "course:user:delete "+"id_course="+s[0]+" id_user="+s[1])
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "courseAdd":
					text := botlogic.SendToMain(chat_id, "course:add")
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "ID новой дисциплины: "+text)
				case "courseDelete":
					text := botlogic.SendToMain(chat_id, "course:del "+"id_course="+args)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "questList":
					s := strings.SplitAfter(args, " ")
					if len(s) != 2 {
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Неверное число аргументов")
						break
					}
					text := botlogic.SendToMain(chat_id, "quest:list:read:other "+"id_user="+s[0]+" id_test="+s[1])
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "questInfo":
					s := strings.SplitAfter(args, " ")
					if len(s) != 2 {
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Неверное число аргументов")
						break
					}
					text := botlogic.SendToMain(chat_id, "quest:read "+"id_test="+s[0]+" id_answer="+s[1])
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "questUpdate":
					s := strings.SplitAfter(args, " ")
					if len(s) != 2 {
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Неверное число аргументов")
						break
					}
					text := botlogic.SendToMain(chat_id, "quest:update "+"id_test="+s[0]+" id_answer="+s[1])
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "questCreate":
					text := botlogic.SendToMain(chat_id, "quest:create "+"id_test="+args)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "questDelete":
					text := botlogic.SendToMain(chat_id, "quest:del "+"id_test="+args)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "testQuestDelete":
					s := strings.SplitAfter(args, " ")
					if len(s) != 2 {
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Неверное число аргументов")
						break
					}
					text := botlogic.SendToMain(chat_id, "test:quest:del "+"id_test="+s[0]+" id_quest="+s[1])
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "testQuestAdd":
					s := strings.SplitAfter(args, " ")
					if len(s) != 2 {
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Неверное число аргументов")
						break
					}
					text := botlogic.SendToMain(chat_id, "test:quest:add "+"id_test="+s[0]+" id_quest="+s[1])
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "testProcedureChange":
					text := botlogic.SendToMain(chat_id, "test:quest:update "+"id_test="+args)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "testAnswerUsersList":
					text := botlogic.SendToMain(chat_id, "test:answer:read "+"id_test="+args)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "testUserAnswers":
					s := strings.SplitAfter(args, " ")
					if len(s) != 2 {
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Неверное число аргументов")
						break
					}
					text := botlogic.SendToMain(chat_id, "test:answer:read:other "+"id_user="+s[0]+" id_test="+s[1])
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "testUserMark":
					s := strings.SplitAfter(args, " ")
					if len(s) != 2 {
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Неверное число аргументов")
						break
					}
					text := botlogic.SendToMain(chat_id, "test:answer:read:other "+"id_user="+s[0]+" id_test="+s[1])
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "attempCreate":
					text := botlogic.SendToMain(chat_id, "test:answer:create "+"id_test="+args)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "attempUpdate":
					text := botlogic.SendToMain(chat_id, "test:answer:update "+"id_test="+args)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "attempEnd":
					text := botlogic.SendToMain(chat_id, "test:answer:del "+"id_test="+args)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "attempRead":
					s := strings.SplitAfter(args, " ")
					if len(s) != 2 {
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Неверное число аргументов")
						break
					}
					text := botlogic.SendToMain(chat_id, "test:answer:read:other "+"id_user="+s[0]+" id_test="+s[1])
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "answerCreate":
					s := strings.SplitAfter(args, " ")
					if len(s) != 2 {
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Неверное число аргументов")
						break
					}
					text := botlogic.SendToMain(chat_id, "answer:create "+"id_test="+s[0]+" id_answer="+s[1])
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "answerRead":
					s := strings.SplitAfter(args, " ")
					if len(s) != 2 {
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Неверное число аргументов")
						break
					}
					text := botlogic.SendToMain(chat_id, "answer:read "+"id_test="+s[0]+" id_answer="+s[1])
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "answerChange":
					s := strings.SplitAfter(args, " ")
					if len(s) != 2 {
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Неверное число аргументов")
						break
					}
					text := botlogic.SendToMain(chat_id, "answer:update "+"id_test="+s[0]+" id_answer="+s[1])
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "answerDelete":
					s := strings.SplitAfter(args, " ")
					if len(s) != 2 {
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Неверное число аргументов")
						break
					}
					text := botlogic.SendToMain(chat_id, "answer:del "+"id_test="+s[0]+" id_answer="+s[1])
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				case "anonme":
					botlogic.SetStatus(chat_id, "Анонимный")
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Статус изменён на Анонимный")
				case "aume":
					botlogic.SetStatus(chat_id, "Авторизованный")
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Статус изменён на Авторизованный")

				default:
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Нет такой команды")
				}

				//отправляем ответ
				bot.Send(msg)

			} else if update.CallbackQuery != nil {
				// обработка нажатия на кнопку
				callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
				if _, err := bot.Request(callback); err != nil {
					log.Panic("Ошибка при обработке callback:", err)
				}

				// Определяем, какая кнопка была нажата
				switch update.CallbackQuery.Data {
				case "button_code":
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID,
						botlogic.Login_type(update.CallbackQuery.Message.Chat.ID))
					bot.Send(msg)
				}
			}
		}
	}
}

func Timer() {
	for {
		log.Println("◘ Timer run ◘")
		newtimer := time.NewTimer(60 * time.Second)

		<-newtimer.C
		log.Println("• Timer stop •")

		// выполняем действия по таймеру
		// проверка входа

		replies := botlogic.Entry()
		for id1, reply := range replies {
			msg := tgbotapi.NewMessage(id1, reply)
			msgs = append(msgs, msg)
		}

		//проверка уведомлений
		notifications := botlogic.Alert()
		for id2, alert := range notifications {
			msg := tgbotapi.NewMessage(id2, alert)
			msgs = append(msgs, msg)
		}
		log.Println("Timer messages: ", msgs)
	}
}