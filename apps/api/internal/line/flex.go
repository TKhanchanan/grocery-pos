package line

import (
	"fmt"
	"time"
)

type Component map[string]any

type Bubble struct {
	Type   string     `json:"type"`
	Size   string     `json:"size,omitempty"`
	Header *Component `json:"header,omitempty"`
	Body   *Component `json:"body,omitempty"`
	Footer *Component `json:"footer,omitempty"`
}

type SaleCompletedInput struct {
	ReceiptNo  string
	Total      float64
	SaleTime   time.Time
	ReceiptURL string
}

type StockAlertInput struct {
	ProductName  string
	LocationName string
	Quantity     int
	ReorderPoint int
}

type TestNotificationInput struct {
	Timestamp time.Time
}

type cardTheme struct {
	HeaderColor   string
	SubtitleColor string
	AccentColor   string
}

var (
	successTheme = cardTheme{HeaderColor: "#009A72", SubtitleColor: "#D1FAE5", AccentColor: "#008060"}
	warningTheme = cardTheme{HeaderColor: "#D97706", SubtitleColor: "#FEF3C7", AccentColor: "#B45309"}
	dangerTheme  = cardTheme{HeaderColor: "#DC2626", SubtitleColor: "#FEE2E2", AccentColor: "#B91C1C"}
	infoTheme    = cardTheme{HeaderColor: "#2563EB", SubtitleColor: "#DBEAFE", AccentColor: "#1D4ED8"}
	neutralTheme = cardTheme{HeaderColor: "#334155", SubtitleColor: "#E2E8F0", AccentColor: "#0F766E"}
)

func BuildSaleCompleted(input SaleCompletedInput) FlexMessage {
	fields := []Component{
		label("เลขที่ใบเสร็จ"),
		value(input.ReceiptNo),
		separator(),
		horizontalValue("ยอดรวม", baht(input.Total), successTheme.AccentColor, "xl"),
	}
	if !input.SaleTime.IsZero() {
		fields = append(fields, horizontalValue("วันที่ขาย", formatTime(input.SaleTime), "#475569", "sm"))
	}

	var footer *Component
	if input.ReceiptURL != "" {
		footer = component(box("vertical", Component{
			"type":   "button",
			"style":  "primary",
			"color":  successTheme.AccentColor,
			"action": Component{"type": "uri", "label": "ดูใบเสร็จ", "uri": input.ReceiptURL},
		}))
	} else {
		footer = component(box("vertical", Component{
			"type":  "text",
			"text":  "บันทึกการขายเรียบร้อยแล้ว",
			"size":  "xs",
			"color": "#64748B",
			"align": "center",
		}))
	}

	return flexCard(
		fmt.Sprintf("ขายสำเร็จ - ยอดรวม %s", baht(input.Total)),
		"ขายสำเร็จ",
		"GroceryPOS",
		successTheme,
		fields,
		footer,
	)
}

func BuildLowStock(input StockAlertInput) FlexMessage {
	fields := stockFields(input, warningTheme.AccentColor)
	fields = append(fields, hint("ควรเติมสต็อกเร็ว ๆ นี้", warningTheme.AccentColor))
	return flexCard(
		fmt.Sprintf("สินค้าใกล้หมด - %s เหลือ %d", input.ProductName, input.Quantity),
		"สินค้าใกล้หมด",
		"GroceryPOS",
		warningTheme,
		fields,
		nil,
	)
}

func BuildOutOfStock(input StockAlertInput) FlexMessage {
	fields := stockFields(input, dangerTheme.AccentColor)
	fields = append(fields, hint("กรุณาเติมสต็อกก่อนขายต่อ", dangerTheme.AccentColor))
	return flexCard(
		fmt.Sprintf("สินค้าหมด - %s", input.ProductName),
		"สินค้าหมด",
		"GroceryPOS",
		dangerTheme,
		fields,
		nil,
	)
}

func BuildReorderPoint(input StockAlertInput) FlexMessage {
	fields := stockFields(input, infoTheme.AccentColor)
	fields = append(fields, hint("ตรวจสอบและสร้างรายการสั่งซื้อ", infoTheme.AccentColor))
	return flexCard(
		fmt.Sprintf("ถึงจุดสั่งซื้อแล้ว - %s", input.ProductName),
		"ถึงจุดสั่งซื้อแล้ว",
		"GroceryPOS",
		infoTheme,
		fields,
		nil,
	)
}

func BuildTestNotification(input TestNotificationInput) FlexMessage {
	fields := []Component{
		label("ข้อความ"),
		value("GroceryPOS LINE notification test"),
		separator(),
		horizontalValue("สถานะ", "เชื่อมต่อสำเร็จ", neutralTheme.AccentColor, "sm"),
	}
	if !input.Timestamp.IsZero() {
		fields = append(fields, horizontalValue("เวลา", formatTime(input.Timestamp), "#475569", "sm"))
	}
	return flexCard(
		"ทดสอบการแจ้งเตือน - เชื่อมต่อสำเร็จ",
		"ทดสอบการแจ้งเตือน",
		"GroceryPOS",
		neutralTheme,
		fields,
		nil,
	)
}

func flexCard(altText, title, subtitle string, theme cardTheme, fields []Component, footer *Component) FlexMessage {
	header := box("vertical",
		Component{"type": "text", "text": title, "weight": "bold", "size": "lg", "color": "#FFFFFF"},
		Component{"type": "text", "text": subtitle, "size": "sm", "color": theme.SubtitleColor},
	)
	header["backgroundColor"] = theme.HeaderColor
	body := box("vertical", fields...)
	body["spacing"] = "md"

	return FlexMessage{
		Type:    "flex",
		AltText: altText,
		Contents: Bubble{
			Type:   "bubble",
			Size:   "mega",
			Header: component(header),
			Body:   component(body),
			Footer: footer,
		},
	}
}

func stockFields(input StockAlertInput, accentColor string) []Component {
	fields := []Component{
		label("สินค้า"),
		value(input.ProductName),
		horizontalValue("สถานที่", defaultLocation(input.LocationName), "#0F172A", "sm"),
		separator(),
		horizontalValue("จำนวนคงเหลือ", fmt.Sprintf("%d", input.Quantity), accentColor, "xl"),
	}
	if input.ReorderPoint > 0 {
		fields = append(fields, horizontalValue("จุดสั่งซื้อ", fmt.Sprintf("%d", input.ReorderPoint), "#475569", "sm"))
	}
	return fields
}

func box(layout string, contents ...Component) Component {
	return Component{"type": "box", "layout": layout, "contents": contents}
}

func label(text string) Component {
	return Component{"type": "text", "text": text, "size": "xs", "color": "#64748B"}
}

func value(text string) Component {
	return Component{"type": "text", "text": text, "weight": "bold", "size": "sm", "wrap": true, "color": "#0F172A"}
}

func separator() Component {
	return Component{"type": "separator", "margin": "md"}
}

func horizontalValue(name, text, color, size string) Component {
	return box("horizontal",
		Component{"type": "text", "text": name, "size": "sm", "color": "#475569", "flex": 2},
		Component{"type": "text", "text": text, "size": size, "weight": "bold", "align": "end", "color": color, "wrap": true, "flex": 3},
	)
}

func hint(text, color string) Component {
	return Component{
		"type":            "box",
		"layout":          "vertical",
		"backgroundColor": "#F8FAFC",
		"cornerRadius":    "md",
		"paddingAll":      "md",
		"contents": []Component{
			{"type": "text", "text": text, "size": "sm", "color": color, "wrap": true},
		},
	}
}

func component(value Component) *Component {
	return &value
}

func baht(value float64) string {
	return fmt.Sprintf("฿%.2f", value)
}

func formatTime(value time.Time) string {
	return value.Format("02/01/2006 15:04")
}

func defaultLocation(value string) string {
	if value == "" {
		return "หน้าร้าน"
	}
	return value
}
