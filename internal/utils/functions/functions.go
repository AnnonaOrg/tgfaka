package functions

import (
	"bytes"
	"encoding/json"
	"errors"
	"image"
	"image/draw"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
	"unicode"

	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/skip2/go-qrcode"
)

func GenerateToken(dataMap map[string]interface{}, secretString string, expireTime time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	for key, value := range dataMap {
		claims[key] = value
	}
	claims["exp"] = time.Now().Add(expireTime).Unix()

	secretByte := []byte(secretString)
	tokenString, err := token.SignedString(secretByte)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
func DecodeToken(tokenString string, secretString string) (map[string]interface{}, error) {
	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) { return []byte(secretString), nil })
	result := make(map[string]interface{})
	if err != nil {
		return result, err
	}
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		for key, value := range claims {
			result[key] = value
		}
		return result, err
	} else {
		return result, err
	}
}

func GenerateRandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

type StructMarshalMode int

const (
	StructToMapIncludeMode StructMarshalMode = iota
	StructToMapExcludeMode
)

// struct转map，适用于值和指针的属性
// 若为指针且值为nil，则跳过不解析。若json标签没有名字，则跳过不解析
// 该情况与上述矛盾，若需要则改代码加输入变量，若为指针且值为nil，则json解析成null
func StructToMap(v interface{}, mode StructMarshalMode, keys ...string) map[string]interface{} {
	resultMap := make(map[string]interface{})
	vValue := reflect.Indirect(reflect.ValueOf(v)) // Automatically handles pointers

	for i := 0; i < vValue.NumField(); i++ {
		field := vValue.Field(i)
		typeField := vValue.Type().Field(i)
		jsonTag := typeField.Tag.Get("json")
		tagParts := strings.Split(jsonTag, ",")
		jsonKey := tagParts[0]

		//跳过值为nil,判断前要判断是否为指针
		//1.用于请求参数，为EditProduct handler在用，同一个edit，用于切换开关和修改内容
		//2.用于返回参数，有这个跳过nil会导致值为nil的kv缺失，如果这里出问题，希望给上order_id:null这种返回值，则函数加一个参数切换
		if field.Kind() == reflect.Ptr && field.IsNil() {
			continue
		}

		//跳过jsongtag为空
		if jsonTag == "" {
			continue
		}
		//排除模式、包括模式
		if mode == StructToMapIncludeMode {
			if !SliceContainString(keys, jsonKey) {
				continue
			}
		} else {
			if SliceContainString(keys, jsonKey) {
				continue
			}
		}

		// 跳过含有omitempty的,但保留输入include为最优先
		if len(tagParts) >= 2 && SliceContainString(tagParts[1:], "omitempty") {
			if !(mode == StructToMapIncludeMode && SliceContainString(keys, jsonKey)) {
				continue
			}
		}

		// 重置uuid默认值,这里是用于返回请求有关外键的参数,有些外键为null,但是默认值是00000
		if field.Type() == reflect.TypeOf(uuid.UUID{}) && field.Interface() == uuid.Nil {
			resultMap[jsonKey] = nil

		} else if field.Type() == reflect.TypeOf(decimal.Decimal{}) {
			resultMap[jsonKey] = field.Interface()

			////
			//value, ok := field.Interface().(string)
			//if ok {
			//	resultMap[jsonKey] = decimal.NewFromString(value)
			//} else {
			//
			//}
		} else {
			resultMap[jsonKey] = field.Interface()
		}
	}

	return resultMap
}

//func StructToMap(v interface{}, mode StructMarshalMode, keys ...string) map[string]interface{} {
//	resultMap := make(map[string]interface{})
//	vValue := reflect.ValueOf(v)
//	vType := vValue.Type()
//
//	//如果结构是指针，取值
//	if vType.Kind() == reflect.Ptr {
//		vValue = vValue.Elem()
//		vType = vValue.Type()
//	}
//
//	for i := 0; i < vValue.NumField(); i++ {
//		field := vValue.Field(i)
//		typeField := vType.Field(i)
//		jsonTag := typeField.Tag.Get("json")
//		tagParts := strings.Split(jsonTag, ",") // 支持多个json值 `json:"fieldname,omitempty"`
//		jsonKey := tagParts[0]
//
//		//跳过值为nil,判断前要判断是否为指针
//		//1.用于请求参数，为EditProduct handler在用，同一个edit，用于切换开关和修改内容
//		//2.用于返回参数，有这个跳过nil会导致值为nil的kv缺失，如果这里出问题，希望给上order_id:null这种返回值，则函数加一个参数切换
//		if field.Kind() == reflect.Ptr && field.IsNil() {
//			continue
//		}
//
//		//跳过jsongtag为空
//		if jsonTag == "" {
//			continue
//		}
//		//排除模式、包括模式
//		if mode == StructToMapIncludeMode {
//			if !SliceContainString(keys, jsonKey) {
//				continue
//			}
//		} else {
//			if SliceContainString(keys, jsonKey) {
//				continue
//			}
//		}
//
//		// 跳过含有omitempty的,但保留输入include为最优先
//		if len(tagParts) >= 2 && SliceContainString(tagParts[1:], "omitempty") {
//			if !(mode == StructToMapIncludeMode && SliceContainString(keys, jsonKey)) {
//				continue
//			}
//		}
//
//		// 重置uuid默认值
//		if field.Type() == reflect.TypeOf(uuid.UUID{}) && field.Interface() == uuid.Nil {
//			resultMap[jsonKey] = nil
//
//		} else {
//			resultMap[jsonKey] = field.Interface()
//		}
//	}
//
//	return resultMap
//}

func MapToStruct(m map[string]interface{}, s interface{}) error {
	if reflect.ValueOf(s).Kind() != reflect.Ptr || reflect.ValueOf(s).Elem().Kind() != reflect.Struct {
		return errors.New("the second argument must be a pointer to a struct")
	}

	jsonData, err := json.Marshal(m)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonData, s)
	if err != nil {
		return err
	}

	return nil
}

// 1 101 convert to {1,50} {51,100} {101,101}
func SplitRangeIntoBulkRanges(start, end, bulkSize int64) []Range {
	if start > end || bulkSize <= 0 {
		// Handle error or invalid input
		return nil
	}

	var bulks []Range

	for i := start; i <= end; i += bulkSize {
		bulkEnd := i + bulkSize - 1
		if bulkEnd > end {
			bulkEnd = end
		}
		bulks = append(bulks, Range{Start: i, End: bulkEnd})
	}

	return bulks
}

type Range struct {
	Start int64
	End   int64
}

func SliceContainString(list []string, a string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
func ParseIDsString(idsString string) ([]uuid.UUID, error) {
	var ids []uuid.UUID

	idStrings := strings.Split(idsString, ",")
	for _, idStr := range idStrings {
		if idStr == "" {
			continue
		}
		id, err := uuid.Parse(idStr)
		if err != nil {
			return ids, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}
func SliceContainDecimal(slice []decimal.Decimal, element decimal.Decimal) bool {
	for _, v := range slice {
		if v.Equal(element) {
			return true
		}
	}
	return false
}

func IsWhitespace(s string) bool {
	for _, r := range s {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}
func GetCurrentSecondTimestampFloat() float64 {
	seconds := float64(time.Now().UnixNano()) / 1e9
	return seconds
}
func GenerateQrCodeBytes(text string) ([]byte, error) {
	// 生成图片
	qrImage, err := qrcode.Encode(text, qrcode.Medium, 256)
	if err != nil {
		return nil, errors.New("生成图片失败")
	}
	img, _, err := image.Decode(bytes.NewReader(qrImage))
	if err != nil {
		return nil, errors.New("读取图片失败")
	}
	// Load a font.
	font, err := LoadFont(GetExecutableDir() + "/static/font.ttf")
	if err != nil {
		return nil, errors.New("读取字体失败")
	}
	// Add label to the image.
	labeledImage, err := AddLabelToImage(img, text, font)
	if err != nil {
		return nil, errors.New("添加文字失败")
	}
	var buf bytes.Buffer
	err = png.Encode(&buf, labeledImage)
	if err != nil {
		return nil, errors.New("编码图片失败")
	}
	return buf.Bytes(), nil
}
func LoadFont(fontFile string) (*truetype.Font, error) {
	// Read the font data.
	fontBytes, err := os.ReadFile(fontFile)
	if err != nil {
		return nil, err
	}
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return nil, err
	}
	return font, nil
}

// AddLabelToImage adds a text label to a given image.
func AddLabelToImage(img image.Image, label string, font *truetype.Font) (image.Image, error) {
	// Set up the image to draw on.
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)

	// Set the text properties.
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(font)
	c.SetFontSize(12) // Adjust font size to your requirements
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(image.Black)

	// Set the position where you want to add the text.
	pt := freetype.Pt(10, 20) // This will need to be adjusted

	// Draw the text.
	_, err := c.DrawString(label, pt)
	if err != nil {
		return nil, err
	}

	return rgba, nil
}
func TruncateToStartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func TruncateToEndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
}
func GetExecutableDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return ""
		//panic(err)
	}

	exeDir := filepath.Dir(exePath)
	return exeDir
}
