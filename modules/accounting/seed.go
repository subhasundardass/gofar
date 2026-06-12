package accounting

import (
	"context"
	"fmt"

	"github.com/subhasundardas/gofar/ent"
	"github.com/subhasundardas/gofar/ent/acct_group"
	"github.com/subhasundardas/gofar/ent/country"
	"github.com/subhasundardas/gofar/ent/state"
)

// Seed inserts default records for the Accounting module.
// It is a no-op if records already exist.
// Called from PostRegister when APP_ENV=development.
func Seed(ctx context.Context, db *ent.Client) error {
	if err := SeedCountries(ctx, db); err != nil {
		return fmt.Errorf("seed countries: %w", err)
	}

	if err := SeedStates(ctx, db); err != nil {
		return fmt.Errorf("seed states: %w", err)
	}

	if err := SeedAccountGroups(ctx, db); err != nil {
		return fmt.Errorf("seed account groups: %w", err)
	}

	if err := SeedLedgers(ctx, db); err != nil {
		return fmt.Errorf("seed ledgers: %w", err)
	}

	return nil
}

// ---------------------------------------------------------------------------
// Countries
// ---------------------------------------------------------------------------

func SeedCountries(ctx context.Context, client *ent.Client) error {
	const (
		countryCode = "IN"
		countryName = "India"
	)

	exists, err := client.Country.
		Query().
		Where(country.CodeEQ(countryCode)).
		Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	_, err = client.Country.
		Create().
		SetName(countryName).
		SetCode(countryCode).
		Save(ctx)
	return err
}

// ---------------------------------------------------------------------------
// States
// ---------------------------------------------------------------------------

type StateSeed struct {
	Name    string
	Code    string
	GSTCode string
}

var IndiaStates = []StateSeed{
	// States
	{"Andhra Pradesh", "AP", "37"},
	{"Arunachal Pradesh", "AR", "12"},
	{"Assam", "AS", "18"},
	{"Bihar", "BR", "10"},
	{"Chhattisgarh", "CG", "22"},
	{"Goa", "GA", "30"},
	{"Gujarat", "GJ", "24"},
	{"Haryana", "HR", "06"},
	{"Himachal Pradesh", "HP", "02"},
	{"Jharkhand", "JH", "20"},
	{"Karnataka", "KA", "29"},
	{"Kerala", "KL", "32"},
	{"Madhya Pradesh", "MP", "23"},
	{"Maharashtra", "MH", "27"},
	{"Manipur", "MN", "14"},
	{"Meghalaya", "ML", "17"},
	{"Mizoram", "MZ", "15"},
	{"Nagaland", "NL", "13"},
	{"Odisha", "OD", "21"},
	{"Punjab", "PB", "03"},
	{"Rajasthan", "RJ", "08"},
	{"Sikkim", "SK", "11"},
	{"Tamil Nadu", "TN", "33"},
	{"Telangana", "TS", "36"},
	{"Tripura", "TR", "16"},
	{"Uttar Pradesh", "UP", "09"},
	{"Uttarakhand", "UK", "05"},
	{"West Bengal", "WB", "19"},

	// Union Territories
	{"Andaman and Nicobar Islands", "AN", "35"},
	{"Chandigarh", "CH", "04"},
	{"Dadra and Nagar Haveli and Daman and Diu", "DN", "26"},
	{"Delhi", "DL", "07"},
	{"Jammu and Kashmir", "JK", "01"},
	{"Ladakh", "LA", "38"},
	{"Lakshadweep", "LD", "31"},
	{"Puducherry", "PY", "34"},
}

func SeedStates(ctx context.Context, client *ent.Client) error {
	india, err := client.Country.
		Query().
		Where(country.CodeEQ("IN")).
		Only(ctx)
	if err != nil {
		return err
	}

	for _, s := range IndiaStates {
		exists, err := client.State.
			Query().
			Where(
				state.CountryIDEQ(india.ID),
				state.NameEQ(s.Name),
			).
			Exist(ctx)
		if err != nil {
			return err
		}
		if exists {
			continue
		}

		_, err = client.State.
			Create().
			SetName(s.Name).
			SetCode(s.Code).
			SetGstCode(s.GSTCode).
			SetCountryID(india.ID).
			Save(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// ---------------------------------------------------------------------------
// Account Groups
// ---------------------------------------------------------------------------

type GroupSeed struct {
	Code       string
	Name       string
	Nature     acct_group.Nature
	ParentCode string
}

var AccountGroups = []GroupSeed{
	// Root
	{"AST", "Assets", acct_group.NatureASSET, ""},
	{"LIA", "Liabilities", acct_group.NatureLIABILITY, ""},
	{"EQT", "Equity", acct_group.NatureEQUITY, ""},
	{"INC", "Income", acct_group.NatureINCOME, ""},
	{"EXP", "Expenses", acct_group.NatureEXPENSE, ""},

	// Assets
	{"CUR_AST", "Current Assets", acct_group.NatureASSET, "AST"},
	{"FIX_AST", "Fixed Assets", acct_group.NatureASSET, "AST"},

	{"CASH", "Cash Accounts", acct_group.NatureASSET, "CUR_AST"},
	{"BANK", "Bank Accounts", acct_group.NatureASSET, "CUR_AST"},
	{"AR", "Accounts Receivable", acct_group.NatureASSET, "CUR_AST"},
	{"INV", "Inventory", acct_group.NatureASSET, "CUR_AST"},
	{"GST_IN", "GST Input", acct_group.NatureASSET, "CUR_AST"},
	{"ADV", "Advances & Deposits", acct_group.NatureASSET, "CUR_AST"},

	{"LAND_GRP", "Land & Building", acct_group.NatureASSET, "FIX_AST"},
	{"FURN", "Furniture & Fixtures", acct_group.NatureASSET, "FIX_AST"},
	{"MACH", "Plant & Machinery", acct_group.NatureASSET, "FIX_AST"},
	{"VEH", "Vehicles", acct_group.NatureASSET, "FIX_AST"},

	// Liabilities
	{"CUR_LIA", "Current Liabilities", acct_group.NatureLIABILITY, "LIA"},
	{"LNG_LIA", "Long Term Liabilities", acct_group.NatureLIABILITY, "LIA"},

	{"AP", "Accounts Payable", acct_group.NatureLIABILITY, "CUR_LIA"},
	{"GST_OUT", "GST Output", acct_group.NatureLIABILITY, "CUR_LIA"},
	{"TDS", "TDS Payable", acct_group.NatureLIABILITY, "CUR_LIA"},
	{"DUTY", "Duties & Taxes", acct_group.NatureLIABILITY, "CUR_LIA"},

	{"SEC_LOAN", "Secured Loans", acct_group.NatureLIABILITY, "LNG_LIA"},
	{"UNSEC_LOAN", "Unsecured Loans", acct_group.NatureLIABILITY, "LNG_LIA"},

	// Equity
	{"CAPITAL", "Capital Account", acct_group.NatureEQUITY, "EQT"},
	{"DRAWING", "Drawings", acct_group.NatureEQUITY, "EQT"},
	{"RETAINED", "Retained Earnings", acct_group.NatureEQUITY, "EQT"},

	// Income
	{"SALES", "Sales", acct_group.NatureINCOME, "INC"},
	{"SERVICE", "Service Income", acct_group.NatureINCOME, "INC"},
	{"OTH_INC", "Other Income", acct_group.NatureINCOME, "INC"},
	{"INT_INC", "Interest Income", acct_group.NatureINCOME, "INC"},

	// Expenses
	{"PURCHASE", "Purchases", acct_group.NatureEXPENSE, "EXP"},
	{"DIR_EXP", "Direct Expenses", acct_group.NatureEXPENSE, "EXP"},
	{"IND_EXP", "Indirect Expenses", acct_group.NatureEXPENSE, "EXP"},
	{"PAYROLL", "Payroll Expenses", acct_group.NatureEXPENSE, "EXP"},
	{"ADMIN", "Administrative Expenses", acct_group.NatureEXPENSE, "EXP"},
	{"FIN_EXP", "Financial Charges", acct_group.NatureEXPENSE, "EXP"},
	{"DEPR", "Depreciation", acct_group.NatureEXPENSE, "EXP"},
}

func SeedAccountGroups(ctx context.Context, client *ent.Client) error {
	exists, err := client.Acct_Group.
		Query().
		Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	groupMap := make(map[string]*ent.Acct_Group, len(AccountGroups))

	for _, g := range AccountGroups {
		create := client.Acct_Group.
			Create().
			SetCode(g.Code).
			SetName(g.Name).
			SetNature(g.Nature).
			SetIsSystem(true)

		if g.ParentCode != "" {
			parent, ok := groupMap[g.ParentCode]
			if !ok {
				return fmt.Errorf(
					"parent group %q not found for %q (declare parents before children)",
					g.ParentCode, g.Code,
				)
			}
			create.SetParent(parent)
		}

		group, err := create.Save(ctx)
		if err != nil {
			return err
		}

		groupMap[g.Code] = group
	}

	return nil
}

// ---------------------------------------------------------------------------
// Ledgers
// ---------------------------------------------------------------------------

type LedgerSeed struct {
	Code      string
	Name      string
	GroupCode string
}

var Ledgers = []LedgerSeed{
	// Cash
	{"CASH_HAND", "Cash In Hand", "CASH"},
	{"PETTY_CASH", "Petty Cash", "CASH"},

	// Receivable
	{"AR_CONTROL", "Accounts Receivable", "AR"},

	// Payable
	{"AP_CONTROL", "Accounts Payable", "AP"},

	// Inventory
	{"INVENTORY", "Inventory Account", "INV"},

	// GST Input
	{"INPUT_CGST", "Input CGST", "GST_IN"},
	{"INPUT_SGST", "Input SGST", "GST_IN"},
	{"INPUT_IGST", "Input IGST", "GST_IN"},

	// GST Output
	{"OUTPUT_CGST", "Output CGST", "GST_OUT"},
	{"OUTPUT_SGST", "Output SGST", "GST_OUT"},
	{"OUTPUT_IGST", "Output IGST", "GST_OUT"},

	// Equity
	{"CAPITAL", "Capital Account", "CAPITAL"},
	{"DRAWING", "Drawings Account", "DRAWING"},
	{"RETAINED", "Retained Earnings", "RETAINED"},

	// Income
	{"SALES", "Sales Account", "SALES"},
	{"SERVICE", "Service Income", "SERVICE"},
	{"OTHER_INCOME", "Other Income", "OTH_INC"},
	{"INTEREST_INCOME", "Interest Income", "INT_INC"},

	// Purchases
	{"PURCHASE", "Purchase Account", "PURCHASE"},

	// Direct Expenses
	{"FREIGHT", "Freight Inward", "DIR_EXP"},
	{"CARRIAGE", "Carriage Inward", "DIR_EXP"},
	{"PACKING", "Packing Charges", "DIR_EXP"},

	// Payroll
	{"SALARY", "Salary Expense", "PAYROLL"},
	{"BONUS", "Bonus Expense", "PAYROLL"},
	{"PF", "PF Contribution", "PAYROLL"},
	{"ESI", "ESI Contribution", "PAYROLL"},

	// Administrative
	{"RENT", "Rent Expense", "ADMIN"},
	{"ELECTRICITY", "Electricity Expense", "ADMIN"},
	{"WATER", "Water Expense", "ADMIN"},
	{"PHONE", "Telephone Expense", "ADMIN"},
	{"INTERNET", "Internet Expense", "ADMIN"},
	{"STATIONERY", "Printing & Stationery", "ADMIN"},
	{"TRAVEL", "Travel Expense", "ADMIN"},
	{"CONVEYANCE", "Conveyance Expense", "ADMIN"},
	{"REPAIR", "Repair & Maintenance", "ADMIN"},

	// Financial
	{"BANK_CHARGE", "Bank Charges", "FIN_EXP"},
	{"INTEREST_EXP", "Interest Expense", "FIN_EXP"},

	// Depreciation
	{"DEPRECIATION", "Depreciation Expense", "DEPR"},

	// Fixed Assets
	{"LAND", "Land", "LAND_GRP"},
	{"BUILDING", "Building", "LAND_GRP"},
	{"FURNITURE", "Furniture", "FURN"},
	{"MACHINERY", "Machinery", "MACH"},
	{"VEHICLE", "Vehicle", "VEH"},
}

func SeedLedgers(ctx context.Context, client *ent.Client) error {
	exists, err := client.Ledger.
		Query().
		Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	groups, err := client.Acct_Group.
		Query().
		All(ctx)
	if err != nil {
		return err
	}

	groupMap := make(map[string]*ent.Acct_Group, len(groups))
	for _, g := range groups {
		groupMap[g.Code] = g
	}

	for _, l := range Ledgers {
		group, ok := groupMap[l.GroupCode]
		if !ok {
			return fmt.Errorf(
				"group %q not found for ledger %q",
				l.GroupCode, l.Code,
			)
		}

		_, err := client.Ledger.
			Create().
			SetCode(l.Code).
			SetName(l.Name).
			SetGroupID(group.ID).
			Save(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
