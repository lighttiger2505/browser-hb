package main

import (
	"database/sql"
	"time"
)

type History struct {
	Title         string
	URL           string
	VisitCount    int
	LastVisitTime time.Time
}

func selectHistory(path string, lastdate string) ([]*History, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	q := `
select
    title,
    url,
    visit_count,
    last_visit_time
from
    urls
order by last_visit_time desc
	`
	if lastdate != "" {
		q = "select title, url, datetime(last_visit_time / 1000000 + (strftime('%s', '1601-01-01')), 'unixepoch') from urls where datetime(last_visit_time / 1000000 + (strftime('%s', '1601-01-01')), 'unixepoch') >= ? order by last_visit_time desc"
	}
	rows, err := db.Query(q, lastdate)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	// 現在のロケールのタイムゾーンを取得
	currentTime := time.Now()
	_, tz := currentTime.Zone()
	histories := []*History{}
	for rows.Next() {
		var lastVisitTime int64
		history := &History{}
		if err := rows.Scan(
			&history.Title,
			&history.URL,
			&history.VisitCount,
			&lastVisitTime,
		); err != nil {
			panic(err)
		}
		// WebKitのタイムスタンプをtime.Timeに変換
		t := webkitToTime(lastVisitTime)
		// PCのタイムゾーンに合わせる
		history.LastVisitTime = t.In(time.FixedZone("Local", tz))

		histories = append(histories, history)
	}
	return histories, nil
}

func webkitToTime(webkitTimestamp int64) time.Time {
	// WebKitの基準日 (1601-01-01 00:00:00 UTC)
	var webkitEpochTime = time.Date(1601, 1, 1, 0, 0, 0, 0, time.UTC)
	// マイクロ秒単位を秒に変換
	seconds := webkitTimestamp / 1000000
	// 基準日にタイムスタンプの経過秒を追加してtime.Timeに変換
	return time.Unix(webkitEpochTime.Unix()+seconds, 0).UTC()
}
