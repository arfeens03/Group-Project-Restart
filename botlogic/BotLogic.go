package botlogic

import (
	"kubete_torrentBot/remote"
	"kubete_torrentBot/strgred"
	"log"
)

func Get_status(chat_id int64) string {

	// получаем статус из бд
	status, _, _ := strgred.Redis_get(chat_id)

	if status == "nil" {
		status = "Неизвестный"
	}

	return status
}

func Get_data(chat_id int64) (string, string, string) {
	return strgred.Redis_get(chat_id)
}

func SetStatus(chat_id int64, neo string) {
	strgred.Redis_delete(chat_id)
	if neo == "Анонимный" {
		strgred.Redis_add2(chat_id, neo, strgred.GenerateEntryToken())
	}
	if neo == "Авторизованный" {
		strgred.Redis_add(chat_id, neo, "fake_acess_token", "fake_update_token")
	}

}

func SendToMain(chat_id int64, code string) string {
	status, acess_token, _ := strgred.Redis_get(chat_id)
	result := code + "\n" + acess_token

	// отправляем result и получаем ответ
	reply := remote.SendMain(result)

	switch reply {
	case "401":

		au_reply := remote.SendAu(result)
		switch au_reply {
		case "401": // токен устарел или не существует
			strgred.Redis_delete(chat_id)
			return "Вы не залогинены! Необходимо авторизоваться:\n/login"
		default: // модуль ответил новой парой токенов
			au_reply = "_\n" + au_reply
			_, new_acess_token, new_update_token := strgred.SplitValue(au_reply)
			strgred.Redis_add(chat_id, status, new_acess_token, new_update_token)
			return SendToMain(chat_id, code)
		}
	case "403":
		return "Недостаточно прав для этого действия!"
	default:
		return reply
	}
}

func Login(chat_id int64) string {
	// достаём токен входа из redis
	_, entry_token, _ := strgred.Redis_get(chat_id)
	// запрос модулю авторизации - проверка токена входа
	code := remote.SendAu(entry_token)
	switch code {
	case "1": // не опознанный/истёкший токен:
		strgred.Redis_delete(chat_id)
		return "Вы не вошли либо время входа истекло."
	case "2": // в доступе отказано:
		strgred.Redis_delete(chat_id)
		return "Неудачная авторизация."
	default: // доступ предоставлен
		// получаем jwt-токен доступа и токен обновления
		code = "_\n" + code
		_, access_token, update_token := strgred.SplitValue(code)
		strgred.Redis_delete(chat_id)
		strgred.Redis_add(chat_id, "Авторизованный", access_token, update_token)
	}
	return ""
}

func Login_type(chat_id int64) string {
	status := Get_status(chat_id)
	switch status {
	case "Неизвестный":
		entry_token := strgred.GenerateEntryToken()
		strgred.Redis_add2(chat_id, status, entry_token)
		// отправляем токен входа модулю авторизации
		// получаем ответ и отправляем
		return remote.SendAu(entry_token)
	case "Анонимный":
		entry_token := strgred.GenerateEntryToken()
		strgred.Redis_delete(chat_id)
		strgred.Redis_add2(chat_id, status, entry_token)
		// отправляем токен входа модулю авторизации
		// получаем ответ и отправляем
		return remote.SendAu(entry_token)
	default:
		log.Panic("Error: Unknown status")
	}
	return ""
}

func Logout(chat_id int64) string {

	// удаляем chat_id из redis
	if !strgred.Redis_delete(chat_id) {
		return "Сеанс уже завершён ранее"
	}
	return "Сеанс завершён."
}

func Logout_all(chat_id int64) string {
	Logout(chat_id)
	// запрос модулю авторизации /logout, отправляем токен обновления
	_, _, update_token := Get_data(chat_id)
	remote.SendAu("/logout" + update_token)
	return "Сеанс завершён на всех устройствах."
}

func Entry() map[int64]string {
	replies := make(map[int64]string)
	chat_ids := strgred.GetSomeIDs("Анонимный")

	for _, user_id0 := range chat_ids {
		user_id := int64(user_id0)
		// мы перебираем анонимных пользователей, поэтому это наверняка entry_token
		_, entry_token, _ := strgred.Redis_get(user_id)
		entry_token += "" // временно
		code := remote.SendAu(entry_token)
		switch code {
		case "1": // неопознаный токен или время действия закончилось
			strgred.Redis_delete(user_id)
		case "2": // в доступе отказано
			strgred.Redis_delete(user_id)
			replies[user_id] = "Статус входа: неудачная авторизация"
		case "3": // доступ предоставлен
			// получаем jwt-токен доступа и токен обновления
			// сохраняем оба токена и статус Авторизованный в базу
			// получаем jwt-токен доступа и токен обновления
			code = "_\n" + code
			_, access_token, update_token := strgred.SplitValue(code)
			strgred.Redis_delete(user_id)
			strgred.Redis_add(user_id, "Авторизованный", access_token, update_token)
			replies[user_id] = "Статус входа: успешная авторизация"
		default:
			log.Panic("Error: Unknown autorization code")
		}
	}
	return replies
}

func Alert() map[int64]string {
	notifications := make(map[int64]string)

	chat_ids := strgred.GetSomeIDs("Авторизованный")

	for _, user_id := range chat_ids {
		// запрос главному модулю на URL /notification по токену доступа
		_, acess_token, _ := strgred.Redis_get(user_id)
		notic := remote.SendMain("/notififcation" + "\n" + acess_token)
		notifications[int64(user_id)] = notic
		// запрос главному модулю на удаление уведомлений по JWT токену доступа
		remote.SendMain("/notififcation /delete" + "\n" + acess_token)
	}
	return notifications
}