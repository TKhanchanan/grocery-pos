package models

import "time"

type Role string

const (
	RoleAdmin   Role = "ADMIN"
	RoleManager Role = "MANAGER"
	RoleCashier Role = "CASHIER"
)

type User struct {
	ID        uint64    `json:"id"`
	Username  string    `json:"username"`
	FullName  string    `json:"fullName"`
	Role      Role      `json:"role"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"createdAt"`
}

type Category struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type Product struct {
	ID           uint64    `json:"id"`
	CategoryID   *uint64   `json:"categoryId"`
	SKU          string    `json:"sku"`
	Barcode      *string   `json:"barcode"`
	Name         string    `json:"name"`
	Unit         string    `json:"unit"`
	Price        float64   `json:"price"`
	Cost         float64   `json:"cost"`
	Threshold    int       `json:"threshold"`
	ReorderPoint int       `json:"reorderPoint"`
	Active       bool      `json:"active"`
	TotalStock   int       `json:"totalStock"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Location struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"createdAt"`
}

type ProductStock struct {
	ProductID    uint64 `json:"productId"`
	LocationID   uint64 `json:"locationId"`
	ProductName  string `json:"productName"`
	SKU          string `json:"sku"`
	LocationName string `json:"locationName"`
	Quantity     int    `json:"quantity"`
}

type StockMovement struct {
	ID             uint64    `json:"id"`
	ProductID      uint64    `json:"productId"`
	LocationID     uint64    `json:"locationId"`
	ReferenceType  string    `json:"referenceType"`
	ReferenceID    *uint64   `json:"referenceId"`
	QuantityChange int       `json:"quantityChange"`
	UnitCost       *float64  `json:"unitCost"`
	Note           string    `json:"note"`
	CreatedBy      *uint64   `json:"createdBy"`
	CreatedAt      time.Time `json:"createdAt"`
	ProductName    string    `json:"productName"`
	LocationName   string    `json:"locationName"`
}

type Alert struct {
	ID           uint64     `json:"id"`
	ProductID    uint64     `json:"productId"`
	LocationID   uint64     `json:"locationId"`
	Type         string     `json:"type"`
	Message      string     `json:"message"`
	ResolvedAt   *time.Time `json:"resolvedAt"`
	CreatedAt    time.Time  `json:"createdAt"`
	ProductName  string     `json:"productName"`
	LocationName string     `json:"locationName"`
	CurrentStock int        `json:"currentStock"`
	Threshold    int        `json:"threshold"`
	ReorderPoint int        `json:"reorderPoint"`
}

type Sale struct {
	ID            uint64     `json:"id"`
	ReceiptNo     string     `json:"receiptNo"`
	LocationID    uint64     `json:"locationId"`
	LocationName  string     `json:"locationName"`
	CashierID     uint64     `json:"cashierId"`
	TotalAmount   float64    `json:"totalAmount"`
	TotalCost     float64    `json:"totalCost"`
	Profit        float64    `json:"profit"`
	PaymentMethod string     `json:"paymentMethod"`
	PaidAmount    float64    `json:"paidAmount"`
	ChangeAmount  float64    `json:"changeAmount"`
	Status        string     `json:"status"`
	CancelledAt   *time.Time `json:"cancelledAt"`
	CreatedAt     time.Time  `json:"createdAt"`
	Items         []SaleItem `json:"items"`
}

type SaleItem struct {
	ID                  uint64  `json:"id"`
	SaleID              uint64  `json:"saleId"`
	ProductID           uint64  `json:"productId"`
	ProductNameSnapshot string  `json:"productNameSnapshot"`
	SKUSnapshot         string  `json:"skuSnapshot"`
	PriceSnapshot       float64 `json:"priceSnapshot"`
	CostSnapshot        float64 `json:"costSnapshot"`
	Quantity            int     `json:"quantity"`
	LineTotal           float64 `json:"lineTotal"`
	LineCost            float64 `json:"lineCost"`
}

type Supplier struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"createdAt"`
}

type PurchaseOrder struct {
	ID         uint64              `json:"id"`
	PONumber   string              `json:"poNumber"`
	SupplierID uint64              `json:"supplierId"`
	Supplier   string              `json:"supplier"`
	LocationID uint64              `json:"locationId"`
	Location   string              `json:"location"`
	Status     string              `json:"status"`
	TotalCost  float64             `json:"totalCost"`
	CreatedAt  time.Time           `json:"createdAt"`
	ReceivedAt *time.Time          `json:"receivedAt"`
	Items      []PurchaseOrderItem `json:"items"`
}

type PurchaseOrderItem struct {
	ID        uint64  `json:"id"`
	POID      uint64  `json:"poId"`
	ProductID uint64  `json:"productId"`
	Product   string  `json:"product"`
	Quantity  int     `json:"quantity"`
	UnitCost  float64 `json:"unitCost"`
	LineCost  float64 `json:"lineCost"`
}

type Setting struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
	FullName string `json:"fullName"`
	Role     Role   `json:"role"`
	Active   bool   `json:"active"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type ProductInput struct {
	CategoryID   *uint64 `json:"categoryId"`
	SKU          string  `json:"sku"`
	Barcode      *string `json:"barcode"`
	Name         string  `json:"name"`
	Unit         string  `json:"unit"`
	Price        float64 `json:"price"`
	Cost         float64 `json:"cost"`
	Threshold    int     `json:"threshold"`
	ReorderPoint int     `json:"reorderPoint"`
	Active       bool    `json:"active"`
}

type StockChangeRequest struct {
	ProductID  uint64  `json:"productId"`
	LocationID uint64  `json:"locationId"`
	Quantity   int     `json:"quantity"`
	UnitCost   float64 `json:"unitCost"`
	Reason     string  `json:"reason"`
}

type StockTransferRequest struct {
	ProductID      uint64 `json:"productId"`
	FromLocationID uint64 `json:"fromLocationId"`
	ToLocationID   uint64 `json:"toLocationId"`
	Quantity       int    `json:"quantity"`
	Reason         string `json:"reason"`
}

type CreateSaleRequest struct {
	LocationID    uint64                  `json:"locationId"`
	PaymentMethod string                  `json:"paymentMethod"`
	PaidAmount    float64                 `json:"paidAmount"`
	Items         []CreateSaleRequestItem `json:"items"`
}

type CreateSaleRequestItem struct {
	ProductID uint64 `json:"productId"`
	Quantity  int    `json:"quantity"`
}

type PurchaseOrderInput struct {
	SupplierID uint64                   `json:"supplierId"`
	LocationID uint64                   `json:"locationId"`
	Items      []PurchaseOrderItemInput `json:"items"`
}

type PurchaseOrderItemInput struct {
	ProductID uint64  `json:"productId"`
	Quantity  int     `json:"quantity"`
	UnitCost  float64 `json:"unitCost"`
}

type ReportSummary struct {
	Revenue      float64 `json:"revenue"`
	Profit       float64 `json:"profit"`
	SalesCount   int     `json:"salesCount"`
	ItemsSold    int     `json:"itemsSold"`
	InventoryVal float64 `json:"inventoryValue"`
	LowAlerts    int     `json:"lowAlerts"`
	OutAlerts    int     `json:"outAlerts"`
}

type ReportRow map[string]any
