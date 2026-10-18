// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package database

import (
	"log"
	"time"

	"github.com/theasli/GamajBot/internal/database/sqlite"
)

type DB struct {
	sql *sqlite.DB
}

func New(filePath string) (*DB, error) {
	sqlDB, err := sqlite.Open("data/bot.db")
	if err != nil {
		return nil, err
	}

	if err := sqlDB.MigrateFromJSON("data/wallets.json"); err != nil {
		log.Printf("⚠️ خطا در مهاجرت از JSON: %v", err)
	}

	return &DB{sql: sqlDB}, nil
}

func (d *DB) Sql() *sqlite.DB {
	return d.sql
}

func (d *DB) Close() error {
	return d.sql.Close()
}

func (d *DB) Save() error {
	return nil
}

func SetCachedRates(dollar, tether int64) {
	sqlite.SetCachedRates(dollar, tether)
}

func GetCachedRates() *sqlite.CurrencyRates {
	return sqlite.GetCachedRates()
}

type Settings = sqlite.Settings
type CurrencyType = sqlite.CurrencyType
type PaymentGatewaySettings = sqlite.PaymentGatewaySettings
type CurrencyRates = sqlite.CurrencyRates

const (
	CurrencyUSD   = sqlite.CurrencyUSD
	CurrencyUSDT  = sqlite.CurrencyUSDT
	CurrencyToman = sqlite.CurrencyToman
	CurrencyRial  = sqlite.CurrencyRial
)

func (d *DB) GetSettings() *Settings {
	s, err := d.sql.GetSettings()
	if err != nil {
		log.Printf("GetSettings: %v", err)
		return &Settings{}
	}
	return s
}

func (d *DB) UpdateSettings(s *Settings) error {
	return d.sql.UpdateSettings(s)
}

func (d *DB) SetSetting(key string, value interface{}) error {
	switch v := value.(type) {
	case string:
		return d.sql.SetSetting(key, v)
	case bool:
		if v {
			return d.sql.SetSetting(key, "true")
		}
		return d.sql.SetSetting(key, "false")
	case int64:
		return d.sql.SetSetting(key, time.Duration(v).String())
	case float64:
		return d.sql.SetSetting(key, time.Duration(int64(v)).String())
	}
	return nil
}

func (d *DB) GetCurrency() CurrencyType {
	return d.sql.GetCurrency()
}

func (d *DB) SetCurrency(c CurrencyType) error {
	return d.sql.SetCurrency(c)
}

func (d *DB) GetPaymentGateways() *PaymentGatewaySettings {
	pg, err := d.sql.GetPaymentGateways()
	if err != nil {
		log.Printf("GetPaymentGateways: %v", err)
		return &PaymentGatewaySettings{}
	}
	return pg
}

func (d *DB) UpdatePaymentGateways(pg *PaymentGatewaySettings) error {
	return d.sql.UpdatePaymentGateways(pg)
}

func (d *DB) SetPaymentGatewayField(key string, value interface{}) error {
	return d.sql.SetPaymentGatewayField(key, value)
}

type User = sqlite.User
type UserInfo = sqlite.UserInfo

func (d *DB) GetOrCreateUser(id int64, firstName, username string) *User {
	u, err := d.sql.GetOrCreateUser(id, firstName, username)
	if err != nil {
		log.Printf("GetOrCreateUser: %v", err)
		return &User{ID: id, FirstName: firstName, Username: username}
	}
	return u
}

func (d *DB) GetUser(id int64) (*User, error) {
	return d.sql.GetUser(id)
}

func (d *DB) GetAllUsers() []UserInfo {
	users, err := d.sql.GetAllUsers()
	if err != nil {
		log.Printf("GetAllUsers: %v", err)
		return nil
	}
	return users
}

func (d *DB) BanUser(id int64, reason string) error {
	return d.sql.BanUser(id, reason)
}

func (d *DB) UnbanUser(id int64) error {
	return d.sql.UnbanUser(id)
}

func (d *DB) IsUserBanned(id int64) bool {
	return d.sql.IsUserBanned(id)
}

func (d *DB) UpdateUserBalance(id int64, delta float64) (float64, error) {
	return d.sql.UpdateUserBalance(id, delta)
}

func (d *DB) GetUserTxCount(userID int64) int {
	txs, err := d.sql.GetHistory(userID, 0)
	if err != nil {
		return 0
	}
	return len(txs)
}

type Transaction = sqlite.Transaction
type TxInfo = sqlite.TxInfo

func (d *DB) CreateCharge(userID int64, payID string, amountUSD float64, amountToman int64, expiresAt time.Time) error {
	return d.sql.CreateCharge(userID, payID, amountUSD, amountToman, expiresAt)
}

func (d *DB) SetTransactionGateway(userID int64, payID, gateway string) {
	_ = d.sql.SetTransactionGateway(userID, payID, gateway)
}

func (d *DB) ConfirmCharge(payID string) (int64, float64, error) {
	return d.sql.ConfirmCharge(payID)
}

func (d *DB) Spend(userID int64, amountUSD float64, description string) (float64, error) {
	return d.sql.Spend(userID, amountUSD, description)
}

func (d *DB) DeductBalance(userID int64, amountUSD float64, description string) (float64, error) {
	return d.sql.DeductBalance(userID, amountUSD, description)
}

func (d *DB) DeductBalanceWithType(userID int64, amountUSD float64, description, txType string) (float64, error) {
	return d.sql.DeductBalanceWithType(userID, amountUSD, description, txType)
}

func (d *DB) GetHistory(userID int64, limit int) ([]Transaction, error) {
	return d.sql.GetHistory(userID, limit)
}

func (d *DB) GetPendingCharge(userID int64) (*Transaction, error) {
	return d.sql.GetPendingCharge(userID)
}

func (d *DB) CancelPending(userID int64) {
	if err := d.sql.CancelPending(userID); err != nil {
		log.Printf("CancelPending: %v", err)
	}
}

func (d *DB) GetAllTransactions() []TxInfo {
	txs, err := d.sql.GetAllTransactions()
	if err != nil {
		log.Printf("GetAllTransactions: %v", err)
		return nil
	}
	return txs
}

type Admin = sqlite.Admin
type AdminRole = sqlite.AdminRole

const (
	RoleSuperAdmin = sqlite.RoleSuperAdmin
	RoleFinance    = sqlite.RoleFinance
	RoleSupport    = sqlite.RoleSupport
	RoleViewer     = sqlite.RoleViewer
)

func (d *DB) GetAdminRole(userID, mainAdminID int64) AdminRole {
	return d.sql.GetAdminRole(userID, mainAdminID)
}

func (d *DB) IsAdmin(userID, mainAdminID int64) bool {
	return d.sql.IsAdmin(userID, mainAdminID)
}

func (d *DB) AddAdmin(userID int64, role AdminRole, deptIDs []int64, addedBy int64) error {
	return d.sql.AddAdmin(userID, role, deptIDs, addedBy)
}

func (d *DB) RemoveAdmin(userID int64) error {
	return d.sql.RemoveAdmin(userID)
}

func (d *DB) SetAdminRole(userID int64, role AdminRole) error {
	return d.sql.SetAdminRole(userID, role)
}

func (d *DB) SetAdminDepts(userID int64, deptIDs []int64) error {
	return d.sql.SetAdminDepts(userID, deptIDs)
}

func (d *DB) GetAllAdmins() []*Admin {
	admins, err := d.sql.GetAllAdmins()
	if err != nil {
		log.Printf("GetAllAdmins: %v", err)
		return nil
	}
	return admins
}

func (d *DB) GetAdminByID(userID int64) *Admin {
	a, err := d.sql.GetAdminByID(userID)
	if err != nil {
		return nil
	}
	return a
}

type Department = sqlite.Department

func (d *DB) AddDepartment(name, emoji, description string) (*Department, error) {
	return d.sql.AddDepartment(name, emoji, description)
}

func (d *DB) RemoveDepartment(id int64) error {
	return d.sql.RemoveDepartment(id)
}

func (d *DB) GetAllDepartments() []*Department {
	depts, err := d.sql.GetAllDepartments()
	if err != nil {
		return nil
	}
	return depts
}

func (d *DB) GetActiveDepartments() []*Department {
	depts, err := d.sql.GetActiveDepartments()
	if err != nil {
		return nil
	}
	return depts
}

func (d *DB) GetDepartmentByID(id int64) *Department {
	dept, err := d.sql.GetDepartmentByID(id)
	if err != nil {
		return nil
	}
	return dept
}

func (d *DB) ToggleDepartment(id int64) error {
	return d.sql.ToggleDepartment(id)
}

type Ticket = sqlite.Ticket
type TicketStatus = sqlite.TicketStatus
type TicketPriority = sqlite.TicketPriority
type TicketMessage = sqlite.TicketMessage
type TicketStats = sqlite.TicketStats

const (
	TicketOpen        = sqlite.TicketOpen
	TicketPending     = sqlite.TicketPending
	TicketAnswered    = sqlite.TicketAnswered
	TicketWaitingUser = sqlite.TicketWaitingUser
	TicketClosed      = sqlite.TicketClosed

	PriorityLow    = sqlite.PriorityLow
	PriorityNormal = sqlite.PriorityNormal
	PriorityHigh   = sqlite.PriorityHigh
	PriorityUrgent = sqlite.PriorityUrgent
)

func (d *DB) CreateTicket(userID int64, firstName, username string, deptID int64, subject, text string, priority TicketPriority) (*Ticket, error) {
	return d.sql.CreateTicket(userID, firstName, username, deptID, subject, text, priority)
}

func (d *DB) GetTicket(id int64) (*Ticket, error) {
	return d.sql.GetTicket(id)
}

func (d *DB) AddTicketMessage(ticketID, fromID int64, isAdmin bool, text string) error {
	return d.sql.AddTicketMessage(ticketID, fromID, isAdmin, text)
}

func (d *DB) SetTicketStatus(id int64, status TicketStatus) error {
	return d.sql.SetTicketStatus(id, status)
}

func (d *DB) CloseTicket(id int64) error {
	return d.sql.CloseTicket(id)
}

func (d *DB) AssignTicket(ticketID, adminID int64) error {
	return d.sql.AssignTicket(ticketID, adminID)
}

func (d *DB) SetTicketPriority(ticketID int64, priority TicketPriority) error {
	return d.sql.SetTicketPriority(ticketID, priority)
}

func (d *DB) TransferTicket(ticketID, newDeptID int64) error {
	return d.sql.TransferTicket(ticketID, newDeptID)
}

func (d *DB) GetUserTickets(userID int64) []*Ticket {
	tickets, err := d.sql.GetUserTickets(userID)
	if err != nil {
		return nil
	}
	return tickets
}

func (d *DB) GetAllTickets(onlyOpen bool) []*Ticket {
	tickets, err := d.sql.GetAllTickets(onlyOpen)
	if err != nil {
		return nil
	}
	return tickets
}

func (d *DB) GetTicketsByDept(deptID int64, onlyOpen bool) []*Ticket {
	tickets, err := d.sql.GetTicketsByDept(deptID, onlyOpen)
	if err != nil {
		return nil
	}
	return tickets
}

func (d *DB) GetTicketStats() *TicketStats {
	s, err := d.sql.GetTicketStats()
	if err != nil {
		return &TicketStats{}
	}
	return s
}

type Category = sqlite.Category
type ProductCategory = sqlite.Category
type Product = sqlite.Product

func (d *DB) CreateCategory(parentID int64, name, emoji string) (*Category, error) {
	return d.sql.CreateCategory(parentID, name, emoji)
}

func (d *DB) GetCategory(id int64) *Category {
	c, err := d.sql.GetCategory(id)
	if err != nil {
		return nil
	}
	return c
}

func (d *DB) GetCategoriesByParent(parentID int64) []*Category {
	cats, err := d.sql.GetCategoriesByParent(parentID)
	if err != nil {
		return nil
	}
	return cats
}

func (d *DB) GetAllCategories() []*Category {
	cats, err := d.sql.GetAllCategories()
	if err != nil {
		return nil
	}
	return cats
}

func (d *DB) DeleteCategory(id int64) error {
	return d.sql.DeleteCategory(id)
}

func (d *DB) ToggleCategory(id int64) error {
	return d.sql.ToggleCategory(id)
}

func (d *DB) CreateProduct(categoryID int64, name, description string, priceUSD float64, deliveryType string) (*Product, error) {
	return d.sql.CreateProduct(categoryID, name, description, priceUSD, deliveryType)
}

func (d *DB) SetProductVPN(id, panelID int64, inboundIDs []int64, trafficGB int64, durationDays int) error {
	return d.sql.SetProductVPN(id, panelID, inboundIDs, trafficGB, durationDays)
}

func (d *DB) GetProduct(id int64) *Product {
	p, err := d.sql.GetProduct(id)
	if err != nil {
		return nil
	}
	return p
}

func (d *DB) GetProductsByCategory(categoryID int64) []*Product {
	prods, err := d.sql.GetProductsByCategory(categoryID)
	if err != nil {
		return nil
	}
	return prods
}

func (d *DB) GetAllProducts() []*Product {
	prods, err := d.sql.GetAllProducts()
	if err != nil {
		return nil
	}
	return prods
}

func (d *DB) DeleteProduct(id int64) error {
	return d.sql.DeleteProduct(id)
}

func (d *DB) ToggleProductAvailable(id int64) (*Product, error) {
	if err := d.sql.ToggleProductAvailable(id); err != nil {
		return nil, err
	}
	return d.GetProduct(id), nil
}

func (d *DB) ToggleProductActive(id int64) error {
	return d.sql.ToggleProductActive(id)
}

func (d *DB) UpdateProductPrice(id int64, priceUSD float64) error {
	return d.sql.UpdateProductPrice(id, priceUSD)
}

type Order = sqlite.Order
type OrderStatus = sqlite.OrderStatus

const (
	OrderPending       = sqlite.OrderPending
	OrderDelivered     = sqlite.OrderDelivered
	OrderRejected      = sqlite.OrderRejected
	OrderCodeRequested = sqlite.OrderCodeRequested
	OrderCodeSent      = sqlite.OrderCodeSent
)

func (d *DB) CreateOrder(userID int64, userName string, productID int64) (*Order, float64, error) {
	return d.sql.CreateOrder(userID, userName, productID)
}

func (d *DB) GetOrder(id string) *Order {
	o, err := d.sql.GetOrder(id)
	if err != nil {
		return nil
	}
	return o
}

func (d *DB) GetOrdersByUser(userID int64) []*Order {
	orders, err := d.sql.GetOrdersByUser(userID)
	if err != nil {
		return nil
	}
	return orders
}

func (d *DB) GetAllOrders() []*Order {
	orders, err := d.sql.GetAllOrders()
	if err != nil {
		return nil
	}
	return orders
}

func (d *DB) UpdateOrder(o *Order) error {
	return d.sql.UpdateOrder(o)
}

func (d *DB) DeliverOrder(id, serviceInfo string) (*Order, error) {
	return d.sql.DeliverOrder(id, serviceInfo)
}

func (d *DB) RejectOrder(id, note string) (*Order, float64, error) {
	return d.sql.RejectOrder(id, note)
}

func (d *DB) RequestOrderCode(id string) (*Order, error) {
	return d.sql.RequestOrderCode(id)
}

func (d *DB) SendOrderCode(id, code string) (*Order, error) {
	return d.sql.SendOrderCode(id, code)
}

type XUIPanel = sqlite.XUIPanel

func (d *DB) AddXUIPanel(p *XUIPanel) (*XUIPanel, error) {
	return d.sql.AddXUIPanel(p)
}

func (d *DB) GetXUIPanel(id int64) *XUIPanel {
	p, err := d.sql.GetXUIPanel(id)
	if err != nil {
		return nil
	}
	return p
}

func (d *DB) GetAllXUIPanels() []*XUIPanel {
	panels, err := d.sql.GetAllXUIPanels()
	if err != nil {
		return nil
	}
	return panels
}

func (d *DB) UpdateXUIPanel(p *XUIPanel) error {
	return d.sql.UpdateXUIPanel(p)
}

func (d *DB) DeleteXUIPanel(id int64) error {
	return d.sql.DeleteXUIPanel(id)
}

func (d *DB) ToggleXUIPanelActive(id int64) error {
	return d.sql.ToggleXUIPanelActive(id)
}

type ManualChargeRequest = sqlite.ManualChargeRequest

var (
	ErrManualChargeNotFound = sqlite.ErrManualChargeNotFound
	ErrManualChargeDone     = sqlite.ErrManualChargeDone
)

func (d *DB) CreateManualChargeRequest(req *ManualChargeRequest) error {
	return d.sql.CreateManualChargeRequest(req)
}

func (d *DB) GetManualChargeRequest(id string) (*ManualChargeRequest, error) {
	return d.sql.GetManualChargeRequest(id)
}

func (d *DB) ApproveManualCharge(id string) (*ManualChargeRequest, float64, error) {
	return d.sql.ApproveManualCharge(id)
}

func (d *DB) RejectManualCharge(id, reason string) (*ManualChargeRequest, error) {
	return d.sql.RejectManualCharge(id, reason)
}

func (d *DB) BlockRejectManualCharge(id, reason string) (*ManualChargeRequest, error) {
	return d.sql.BlockRejectManualCharge(id, reason)
}

func (d *DB) MarkManualChargeNeedsReview(id, note string) (*ManualChargeRequest, error) {
	return d.sql.MarkManualChargeNeedsReview(id, note)
}

func (d *DB) SetReceipt(id, fileID, receiptType string) error {
	return d.sql.SetReceipt(id, fileID, receiptType)
}

func (d *DB) FindManualChargeByPayID(payID string) *ManualChargeRequest {
	r, err := d.sql.FindManualChargeByPayID(payID)
	if err != nil {
		return nil
	}
	return r
}

func (d *DB) UpdateManualCharge(req *ManualChargeRequest) error {
	return d.sql.UpdateManualCharge(req)
}

type Stats = sqlite.Stats
type SalesStats = sqlite.SalesStats
type ProductSales = sqlite.ProductSales

func (d *DB) GetStats() *Stats {
	s, err := d.sql.GetStats()
	if err != nil {
		return &Stats{}
	}
	return s
}

func (d *DB) GetSalesStats(since time.Time) *SalesStats {
	s, err := d.sql.GetSalesStats(since)
	if err != nil {
		return &SalesStats{}
	}
	return s
}

func (d *DB) GetTopProducts(since time.Time, limit int) []ProductSales {
	ps, err := d.sql.GetTopProducts(since, limit)
	if err != nil {
		return nil
	}
	return ps
}

type SubscriptionPlan = sqlite.SubscriptionPlan
type DynamicPrice = sqlite.DynamicPrice
type SubscriptionType = sqlite.SubscriptionType

const (
	SubTypeFixed   = sqlite.SubTypeFixed
	SubTypeDynamic = sqlite.SubTypeDynamic
)

func (d *DB) CreateSubscriptionPlan(p *SubscriptionPlan) (*SubscriptionPlan, error) {
	return d.sql.CreateSubscriptionPlan(p)
}

func (d *DB) GetSubscriptionPlan(id int64) (*SubscriptionPlan, error) {
	return d.sql.GetSubscriptionPlan(id)
}

func (d *DB) GetActiveSubscriptionPlans() ([]*SubscriptionPlan, error) {
	return d.sql.GetActiveSubscriptionPlans()
}

func (d *DB) GetAllSubscriptionPlans() ([]*SubscriptionPlan, error) {
	return d.sql.GetAllSubscriptionPlans()
}

func (d *DB) DeleteSubscriptionPlan(id int64) error {
	return d.sql.DeleteSubscriptionPlan(id)
}

func (d *DB) ToggleSubscriptionPlan(id int64) error {
	return d.sql.ToggleSubscriptionPlan(id)
}

func (d *DB) CreateDynamicPrice(dp *DynamicPrice) (*DynamicPrice, error) {
	return d.sql.CreateDynamicPrice(dp)
}

func (d *DB) GetDynamicPrices() ([]*DynamicPrice, error) {
	return d.sql.GetDynamicPrices()
}

func (d *DB) FindDynamicPrice(trafficGB int64, durationDays int, limitIP int) (*DynamicPrice, error) {
	return d.sql.FindDynamicPrice(trafficGB, durationDays, limitIP)
}

func (d *DB) GetDynamicPrice(id int64) (*DynamicPrice, error) {
	return d.sql.GetDynamicPrice(id)
}

func (d *DB) DeleteDynamicPrice(id int64) error {
	return d.sql.DeleteDynamicPrice(id)
}

// ─── Category Extra ───
type DeliveryField = sqlite.DeliveryField
type ServiceAction = sqlite.ServiceAction

func (d *DB) GetCategoryDeliveryFields(catID int64) ([]DeliveryField, error) {
	return d.sql.GetCategoryDeliveryFields(catID)
}

func (d *DB) SetCategoryDeliveryFields(catID int64, fields []DeliveryField) error {
	return d.sql.SetCategoryDeliveryFields(catID, fields)
}

func (d *DB) GetCategoryServiceActions(catID int64) ([]ServiceAction, error) {
	return d.sql.GetCategoryServiceActions(catID)
}

func (d *DB) SetCategoryServiceActions(catID int64, actions []ServiceAction) error {
	return d.sql.SetCategoryServiceActions(catID, actions)
}

func (d *DB) AddCategoryDeliveryField(catID int64, f DeliveryField) error {
	return d.sql.AddCategoryDeliveryField(catID, f)
}

func (d *DB) RemoveCategoryDeliveryField(catID int64, key string) error {
	return d.sql.RemoveCategoryDeliveryField(catID, key)
}

func (d *DB) AddCategoryServiceAction(catID int64, a ServiceAction) error {
	return d.sql.AddCategoryServiceAction(catID, a)
}

func (d *DB) RemoveCategoryServiceAction(catID int64, key string) error {
	return d.sql.RemoveCategoryServiceAction(catID, key)
}

// ─── Order Actions ───
type ActionRequest = sqlite.ActionRequest
type ActionRequestStatus = sqlite.ActionRequestStatus

const (
	ActionPending    = sqlite.ActionPending
	ActionInProgress = sqlite.ActionInProgress
	ActionDone       = sqlite.ActionDone
	ActionRejected   = sqlite.ActionRejected
)

func (d *DB) GetOrderDeliveryData(orderID string) (map[string]string, error) {
	return d.sql.GetOrderDeliveryData(orderID)
}

func (d *DB) SetOrderDeliveryData(orderID string, data map[string]string) error {
	return d.sql.SetOrderDeliveryData(orderID, data)
}

func (d *DB) GetOrderActionRequests(orderID string) ([]ActionRequest, error) {
	return d.sql.GetOrderActionRequests(orderID)
}

func (d *DB) AddOrderActionRequest(orderID string, req ActionRequest) error {
	return d.sql.AddOrderActionRequest(orderID, req)
}

func (d *DB) UpdateOrderActionStatus(orderID, reqID string, status ActionRequestStatus, adminID int64, note string) error {
	return d.sql.UpdateOrderActionStatus(orderID, reqID, status, adminID, note)
}

func (d *DB) GetOrderActionRequest(orderID, reqID string) (*ActionRequest, error) {
	return d.sql.GetOrderActionRequest(orderID, reqID)
}

type LogTopic = sqlite.LogTopic

var ErrLogTopicNotFound = sqlite.ErrLogTopicNotFound

func (d *DB) SetLogTopic(topicKey string, threadID int64, title string) error {
	return d.sql.SetLogTopic(topicKey, threadID, title)
}

func (d *DB) GetLogTopic(topicKey string) (*LogTopic, error) {
	return d.sql.GetLogTopic(topicKey)
}

func (d *DB) GetAllLogTopics() ([]*LogTopic, error) {
	return d.sql.GetAllLogTopics()
}

func (d *DB) DeleteLogTopic(topicKey string) error {
	return d.sql.DeleteLogTopic(topicKey)
}

func (d *DB) DeleteAllLogTopics() error {
	return d.sql.DeleteAllLogTopics()
}

func (d *DB) CreatePlanOrder(userID int64, userName string, planName string, priceUSD float64, trafficGB int64, durationDays int, limitIP int, panelID int64, inboundIDs []int64) (*Order, float64, error) {
	return d.sql.CreatePlanOrder(userID, userName, planName, priceUSD, trafficGB, durationDays, limitIP, panelID, inboundIDs)
}

type DailyRevenue = sqlite.DailyRevenue
type MonthlyComparison = sqlite.MonthlyComparison

func (d *DB) GetDailyRevenue(days int) ([]DailyRevenue, error) {
	return d.sql.GetDailyRevenue(days)
}

func (d *DB) GetMonthlyComparison() (*MonthlyComparison, error) {
	return d.sql.GetMonthlyComparison()
}

func (d *DB) GetRevenueByGateway(days int) (map[string]float64, error) {
	return d.sql.GetRevenueByGateway(days)
}

func (d *DB) SetUserLanguage(id int64, lang string) error {
	return d.sql.SetUserLanguage(id, lang)
}
