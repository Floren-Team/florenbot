package helpers

import (
	"florenbot/engine/structs"
	engine "florenbot/engine"
	"log"
)

func CreateReport(user_id uint64, text string) error {
	_, err := engine.DB.Exec("INSERT INTO reports (user_id, text) VALUES (?, ?)", user_id, text)
	if err != nil {
		log.Println(err)
	}
	return err
}

func GetReports() []structs.Report {
	var reports []structs.Report
	rows, err := engine.DB.Query("SELECT user_id, text FROM reports")
	if err != nil {
		log.Println(err)
	}
	for rows.Next() {
		var report structs.Report
		err := rows.Scan(&report.AngryId, &report.UserId, &report.Text)
		if err != nil {
			log.Println(err)
		}
		reports = append(reports, report)
	}
	return reports
}

func HasReport(user_id uint64) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM reports WHERE user_id = ?)"

	err := engine.DB.QueryRow(query, user_id).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func DeleteReport(user_id uint64) error {
	_, err := engine.DB.Exec("DELETE FROM reports WHERE user_id = ?", user_id)
	if err != nil {
		return err
	}

	return nil
}
