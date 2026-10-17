package dto

type Serie struct {
    Date    string `json:"date"`
    Income  int64  `json:"income"`
    Expense int64  `json:"expense"`
}

type SummaryResponse struct {
    TotalIncome  int64   `json:"total_income"`
    TotalExpense int64   `json:"total_expense"`
    Balance      int64   `json:"balance"`
    Series       []Serie `json:"series"`
}