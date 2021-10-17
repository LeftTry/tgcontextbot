package storage

import (
	"context"
	"fmt"
	db "github.com/jackc/pgx/v4"
	"strings"
	handle "tgcontextbot/internal/handling"
)

func AddWordToBlacklist(idd int64, badWord string) error {
	conn, err := db.Connect(context.Background(), "postgres://postgres:password@localhost:5432/test")

	if err != nil {
		handle.HandleError(err)
	}

	defer func(conn *db.Conn, ctx context.Context) {
		err := conn.Close(ctx)
		if err != nil {
			handle.HandleError(err)
		}
	}(conn, context.Background())

	var allBadWords = strings.Split(badWord, " ")

	var badWordByChat []string = nil

	newRows, Err := conn.Query(context.Background(), "select badword from added_to_chats where id = $1", idd)

	if Err != nil {
		return Err
	}

	if newRows == nil {
		return nil
	}

	for newRows.Next() {
		var txt string
		errWithParse := newRows.Scan(&txt)
		if errWithParse != nil {
			handle.HandleError(errWithParse)
		}
		badWordByChat = append(badWordByChat, txt)
	}

	var flag = false
	var i = 0
	var j = 0
	for i = 0; i < len(allBadWords); i++ {
		flag = false
		for j = 0; j < len(badWordByChat); j++ {
			if allBadWords[i] == badWordByChat[j] {
				flag = true
			}
		}

		if !flag {
			_, err = conn.Exec(context.Background(), "insert into added_to_chats (id, badword) values ($1, $2)", idd, allBadWords[i])

			if err != nil {
				handle.HandleError(err)
			}
		}
	}

	return nil
}

func GetAllBadWordsByChat(idd int64) (*[]string, error) {
	conn, err := db.Connect(context.Background(), "postgres://postgres:password@localhost:5432/test")

	if err != nil {
		handle.HandleError(err)
		return nil, err
	}

	defer func(conn *db.Conn, ctx context.Context) {
		err := conn.Close(ctx)
		if err != nil {
			handle.HandleError(err)
		}
	}(conn, context.Background())

	newRows, Err := conn.Query(context.Background(), "select badword from added_to_chats where id = $1", idd)

	if Err != nil {
		handle.HandleError(Err)
		return nil, Err
	}

	if newRows == nil {
		return nil, nil
	}

	var allWords []string = nil

	for newRows.Next() {
		var txt string

		errWithParse := newRows.Scan(&txt)

		if errWithParse != nil {
			handle.HandleError(errWithParse)
			return nil, errWithParse
		}

		allWords = append(allWords, txt)

	}

	return &allWords, nil

}

func DeleteWordFromBlacklist(idd int64, badWord string) error {
	conn, err := db.Connect(context.Background(), "postgres://postgres:password@localhost:5432/test")

	if err != nil {
		handle.HandleError(err)
	}

	defer func(conn *db.Conn, ctx context.Context) {
		err := conn.Close(ctx)
		if err != nil {
			handle.HandleError(err)
		}
	}(conn, context.Background())

	var allBadWords = strings.Split(badWord, " ")

	var badWordByChat []string = nil

	newRows, Err := conn.Query(context.Background(), "select badword from added_to_chats where id = $1", idd)

	if Err != nil {
		handle.HandleError(Err)
	}

	if newRows == nil {
		return nil
	}

	for newRows.Next() {
		var txt string
		errWithParse := newRows.Scan(&txt)
		if errWithParse != nil {
			return errWithParse
		}
		badWordByChat = append(badWordByChat, txt)
	}

	var flag = false
	var i = 0
	var j = 0
	for i = 0; i < len(allBadWords); i++ {
		flag = true
		for j = 0; j < len(badWordByChat); j++ {
			if allBadWords[i] == badWordByChat[j] {
				flag = false
			}
		}

		if !flag {
			_, err = conn.Exec(context.Background(), "delete from added_to_chats where id = $1 and badword = $2", idd, allBadWords[i])

			if err != nil {
				handle.HandleError(err)
			}
		}
	}

	return nil
}

func AddExceptionToChat(idd int64, excepted string, badword string) (error, string) {
	conn, err := db.Connect(context.Background(), "postgres://postgres:password@localhost:5432/test")

	if err != nil {
		handle.HandleError(err)
		return err, "Мы не сумели установить соединение с базой данных."
	}

	defer func(conn *db.Conn, ctx context.Context) {
		err := conn.Close(ctx)
		if err != nil {
			handle.HandleError(err)
		}
	}(conn, context.Background())

	newRows, Err := conn.Query(context.Background(), "select autoinc_id from except_from_bad_words where bad_word = $1 and username = $2 and chat_id = $3", badword, excepted, idd)

	var cnt int64

	if Err != nil {
		handle.HandleError(Err)
		return Err, "Ошибка на сервере. Пожалуйста, попробуйте снова через некоторое время."
	}

	for newRows.Next() && cnt < 2 {
		cnt++
	}

	if cnt != 0 {
		return nil, "Уже добавлено."
	}

	_, Err = conn.Exec(context.Background(), "insert into except_from_bad_words (bad_word, username, chat_id) values($1, $2, $3)", badword, excepted, idd)

	if Err != nil {
		return Err, "Ошибка на сервере. Пожалуйста, попробуйте снова через некоторое время."
	}

	return nil, "Успешно добавили исключение в базу данных."

}

func GetExceptionsByUsername(idd int64, excepted string) (*[]string, error, string) {
	conn, err := db.Connect(context.Background(), "postgres://postgres:password@localhost:5432/test")

	if err != nil {
		return nil, err, "Мы не сумели установить соединение с базой данных."
	}

	defer func(conn *db.Conn, ctx context.Context) {
		err := conn.Close(ctx)
		if err != nil {
			handle.HandleError(err)
		}
	}(conn, context.Background())

	newRows, Err := conn.Query(context.Background(), "select bad_word from except_from_bad_words where username = $1 and chat_id = $2", excepted, idd)

	if Err != nil {
		handle.HandleError(Err)
		return nil, Err, "Ошибка на сервере. Пожалуйста, повторите попытку позже."
	}

	var ret1 []string
	var scn string

	for newRows.Next() {
		Err = newRows.Scan(&scn)

		if Err != nil {
			handle.HandleError(Err)
			return nil, Err, "Ошибка на сервере. Пожалуйста, повторите попытку позже."
		}

		ret1 = append(ret1, scn)
	}

	return &ret1, nil, "Все успешно."
}

func GetAllExceptionsByChat(idd int64) (*[]string, *[]string, error, string) {
	conn, err := db.Connect(context.Background(), "postgres://postgres:password@localhost:5432/test")

	if err != nil {
		return nil, nil, err, "Мы не сумели установить соединение с базой данных."
	}

	defer func(conn *db.Conn, ctx context.Context) {
		err := conn.Close(ctx)
		if err != nil {
			handle.HandleError(err)
		}
	}(conn, context.Background())

	newRows, Err := conn.Query(context.Background(), "select bad_word, username from except_from_bad_words where chat_id = $1", idd)

	if Err != nil {
		return nil, nil, Err, "На сервере произошла ошибка. Пожалуйста, попробуйте снова через некоторое время."
	}

	var ret1 []string = nil
	var ret2 []string = nil
	var scn1 = ""
	var scn2 = ""

	for newRows.Next() {
		Err = newRows.Scan(&scn1, &scn2)
		if Err != nil {
			return nil, nil, Err, "Произошла ошибка при работе с базой данных."
		}
		ret1 = append(ret1, scn1)
		ret2 = append(ret2, scn2)
	}

	return &ret1, &ret2, nil, "Все успешно."
}

func DeleteExceptionFromChat(idd int64, excepted string, badword string) (error, string) {
	fmt.Println(excepted, "\n", badword)
	conn, err := db.Connect(context.Background(), "postgres://postgres:password@localhost:5432/test")

	if err != nil {
		handle.HandleError(err)
		return err, "Мы не сумели установить соединение с базой данных."
	}

	newRows, Err := conn.Query(context.Background(), "select autoinc_id from except_from_bad_words where bad_word = $1 and username = $2 and chat_id = $3", badword, excepted, idd)

	if Err != nil {
		return Err, "Ошибка на сервере. Пожалуйста, попробуйте снова через некоторое время."
	}

	var cnt int64 = 0
	for newRows.Next() && cnt < 2 {
		cnt++
	}
	if cnt == 0 {
		return nil, "Такого исключения не существует!"
	}

	_, Err = conn.Exec(context.Background(), "delete from except_from_bad_words where bad_word = $1 and username = $2 and chat_id = $3", badword, excepted, idd)

	if Err != nil {
		return Err, "Ошибка на сервере. Пожалуйста, попробуйте снова через некоторое время."
	}

	return nil, "Все успешно."
}
