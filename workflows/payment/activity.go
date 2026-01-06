package payment

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/alifpay/temporal/infra/rmq"
	"github.com/alifpay/temporal/models"
)

//https://nbt.tj/ru/kurs/export_xml_dynamic.php?d1=2025-01-20&d2=2025-01-20&cn=840&cs=USD&export=xml

/*
<?xml version="1.0" encoding="windows-1251" ?>
<ValCurs ID="840"  DateRange1="20/01/2025" DateRange2="20/01/2025"  Name="">
<Record Date="20.01.2025" Id="840">
  <CharCode>USD</CharCode>
     <Nominal>1</Nominal>
   <Value>10.9505</Value>
</Record>
</ValCurs>

*/

type ValCurs struct {
	XMLName xml.Name `xml:"ValCurs"`
	Records []Record `xml:"Record"`
}

type Record struct {
	Date     string `xml:"Date,attr"`
	Id       string `xml:"Id,attr"`
	CharCode string `xml:"CharCode"`
	Nominal  int    `xml:"Nominal"`
	Value    string `xml:"Value"`
}

func GetUSDRate(ctx context.Context, day time.Time) (float64, error) {
	date := day.Format("2006-01-02")
	url := fmt.Sprintf("https://nbt.tj/ru/kurs/export_xml_dynamic.php?d1=%s&d2=%s&cn=840&cs=USD&export=xml", date, date)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("bad status: %s", resp.Status)
	}

	decoder := xml.NewDecoder(resp.Body)
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		if charset == "windows-1251" {
			return input, nil
		}
		return nil, fmt.Errorf("unknown charset: %s", charset)
	}

	var valCurs ValCurs
	if err := decoder.Decode(&valCurs); err != nil {
		return 0, err
	}

	if len(valCurs.Records) == 0 {
		return 0, fmt.Errorf("no records found")
	}

	valStr := valCurs.Records[0].Value
	rate, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return 0, err
	}

	return rate, nil
}

func SendPayment(ctx context.Context, p models.Payment) error {
	js, err := json.Marshal(p)
	if err != nil {
		return err
	}
	err = rmq.Publish(ctx, "test.key", js)
	if err != nil {
		return err
	}
	return nil
}

func UpdatePaymentStatus(ctx context.Context, status models.PaymentStatus) error {
	// This is a placeholder for updating payment status logic
	log.Println("Updating payment status:", status)
	return nil
}
