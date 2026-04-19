package main

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	iconv "github.com/djimenez/iconv-go"
	"github.com/gin-gonic/gin"
	"github.com/gocarina/gocsv"
	"github.com/yeka/zip"
)

const (
	// EncodeTypeUTF8 - CSV File Encoding(utf-8)
	EncodeTypeUTF8 = "utf-8"
	// EncodeTypeSJIS - CSV File Encoding(ShiftJIS)
	EncodeTypeSJIS = "sjis"
)

// TestData - CSVデータ構造体
type TestData struct {
	Name       string    `csv:"name"`
	Phonetic   string    `csv:"phonetic"`
	Address    string    `csv:"address"`
	Birthday   string    `csv:"birthday"`
	CretedDate time.Time `csv:"cretedDate"`
	Dummy      string    `csv:"-"`
}

func main() {
	ginEngine := gin.Default()
	ginEngine.GET("/test", csvDL)

	ginEngine.Run()
}

func csvDL(c *gin.Context) {

	// CSVファイル文字コード指定
	encodingType := EncodeTypeUTF8
	// Zipファイルパスワード有無設定
	isPassword := false

	// CSV・ZIPファイル名設定
	fileName := getDefaultFileName()
	csvFileName := fmt.Sprintf("%s.csv", fileName)
	zipFileName := fmt.Sprintf("%s.zip", fileName)

	// CSVデータ
	nowDt := time.Now()
	csvdataList := []TestData{}
	csvdataList = append(csvdataList, TestData{Name: "テスト太郎", Phonetic: "テストタロウ", Address: "東京都千代田区大手町１−２−３", Birthday: "2000/01/01", CretedDate: nowDt})
	csvdataList = append(csvdataList, TestData{Name: "テスト二郎", Phonetic: "テストジロウ", Address: "東京都千代田区大手町１−２−３", Birthday: "2000/01/01", CretedDate: nowDt, Dummy: "sss"})
	csvdataList = append(csvdataList, TestData{Name: "テスト三郎", Phonetic: "テストサブロウ", Address: "東京都千代田区大手町１−２−３", Birthday: "2000/01/01", CretedDate: nowDt, Dummy: "sss"})

	csvDatab, err := getCsvData(csvdataList, encodingType)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	fmt.Printf("csvdata size = %d\n", binary.Size(csvDatab))

	zipBuffer, err := compress(csvFileName, csvDatab, isPassword)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	fmt.Printf("zipdata size = %d\n", binary.Size(zipBuffer.Bytes()))

	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", zipFileName))
	c.Header("Content-Type", "application/zip")
	c.Writer.Write(zipBuffer.Bytes())
	c.Abort()
}

func getDefaultFileName() string {
	nowDt := time.Now()
	// 年月日フォーマット：yyyyMMddHHmmss
	return nowDt.Format("20060102150405")
}

// CSVデータ生成
func getCsvData(f interface{}, encodeType string) ([]byte, error) {
	var (
		_bytes []byte
		err    error
	)
	// CSVWriterオプション設定
	gocsv.SetCSVWriter(func(out io.Writer) *gocsv.SafeCSVWriter {
		var writer io.Writer
		switch encodeType {
		case EncodeTypeUTF8:
			// utf-8
			writer = out
		case EncodeTypeSJIS:
			// shift-jis
			writer, _ = iconv.NewWriter(out, EncodeTypeUTF8, EncodeTypeSJIS)
		}
		csvWriter := csv.NewWriter(writer)
		csvWriter.Comma = ','
		csvWriter.UseCRLF = true
		return gocsv.NewSafeCSVWriter(csvWriter)
	})

	var dataList []TestData
	switch cast := f.(type) {
	case []TestData:
		dataList = cast
		// Byte単位のCSVデータを作成する
		_bytes, err = gocsv.MarshalBytes(&dataList)
	default:
		return nil, errors.New("サポート外の構造体のため、変換できません")
	}

	if err != nil {
		return nil, err
	}
	return _bytes, nil
}

// ZIPファイルデータ生成
func compress(csvFileName string, csvData []byte, isPassword bool) (*bytes.Buffer, error) {

	var (
		writer io.Writer
		err    error
	)

	password := "12345"

	// ZipWriterの生成
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	defer zipWriter.Close()

	// FileHeaderの作成
	if isPassword {
		// パスワード有り
		writer, err = zipWriter.Encrypt(csvFileName, password, zip.StandardEncryption)
	} else {
		// パスワード無し
		writer, err = zipWriter.Create(csvFileName)
	}

	if err != nil {
		return nil, err
	}

	// CSVデータの書き込み
	writer.Write(csvData)

	return buf, nil
}
