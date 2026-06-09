package line

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestBuildSaleCompleted(t *testing.T) {
	message := BuildSaleCompleted(SaleCompletedInput{
		ReceiptNo: "RC20260609170232703354",
		Total:     45,
		SaleTime:  time.Date(2026, 6, 9, 17, 2, 32, 0, time.FixedZone("ICT", 7*60*60)),
	})

	body, err := json.Marshal(message)
	if err != nil {
		t.Fatal(err)
	}
	got := string(body)
	for _, want := range []string{`"type":"flex"`, `"ขายสำเร็จ"`, `"RC20260609170232703354"`, `"฿45.00"`, `"เลขที่ใบเสร็จ"`} {
		if !strings.Contains(got, want) {
			t.Errorf("payload missing %s: %s", want, got)
		}
	}
}

func TestBuildStockCards(t *testing.T) {
	input := StockAlertInput{
		ProductName:  "ข้าวหอมมะลิถุง 1 กก.",
		LocationName: "หน้าร้าน",
		Quantity:     0,
		ReorderPoint: 5,
	}
	tests := []struct {
		name    string
		message FlexMessage
		want    string
	}{
		{name: "low stock", message: BuildLowStock(input), want: "สินค้าใกล้หมด"},
		{name: "out of stock", message: BuildOutOfStock(input), want: "สินค้าหมด"},
		{name: "reorder point", message: BuildReorderPoint(input), want: "ถึงจุดสั่งซื้อแล้ว"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.message)
			if err != nil {
				t.Fatal(err)
			}
			got := string(body)
			for _, want := range []string{tt.want, input.ProductName, input.LocationName, "จำนวนคงเหลือ", "จุดสั่งซื้อ"} {
				if !strings.Contains(got, want) {
					t.Errorf("payload missing %s: %s", want, got)
				}
			}
		})
	}
}

func TestBuildTestNotification(t *testing.T) {
	message := BuildTestNotification(TestNotificationInput{Timestamp: time.Now()})
	body, err := json.Marshal(message)
	if err != nil {
		t.Fatal(err)
	}
	got := string(body)
	for _, want := range []string{"ทดสอบการแจ้งเตือน", "GroceryPOS LINE notification test", "เชื่อมต่อสำเร็จ"} {
		if !strings.Contains(got, want) {
			t.Errorf("payload missing %s: %s", want, got)
		}
	}
}
