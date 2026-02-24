package geo

import (
	"net"
	"sync"

	"github.com/oschwald/geoip2-golang"
)

var (
	db      *geoip2.Reader
	once    sync.Once
	initErr error
)

// InitGeoDB открывает базу GeoIP. Должна вызываться при старте приложения.
func InitGeoDB(path string) error {
	once.Do(func() {
		db, initErr = geoip2.Open(path)
	})
	return initErr
}

// GetCountryCode возвращает двухбуквенный код страны для IP-адреса.
// Если определить не удалось, возвращается пустая строка.
func GetCountryCode(ipStr string) string {
	if db == nil {
		return ""
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return ""
	}
	record, err := db.Country(ip)
	if err != nil {
		return ""
	}
	return record.Country.IsoCode
}
