package canonical_model

import (
	"time"

	"league.com/rulemaker/msg"
)

type State string

const (
	StateTerminated State = "terminated"
	StateActive     State = "active"
	StateIneligible State = "ineligible"
)

type Benefit struct {
	BenefitType   string    `json:"benefit_type" bson:"benefit_type"`
	Status        State     `json:"status" bson:"status"`
	EffectiveDate time.Time `json:"effective_date" bson:"effective_date"`
}

type Benefits []Benefit

type Dependent struct {
	FirstName                 string     `json:"first_name" bson:"first_name"`
	LastName                  string     `json:"last_name" bson:"last_name"`
	DateOfBirth               *time.Time `json:"date_of_birth" bson:"date_of_birth"`
	Sex                       string     `json:"sex" bson:"sex"` // gender is going to be replaced by Sex soon
	Relationship              string     `json:"relationship" bson:"relationship"`
	RelationshipEffectiveDate *time.Time `json:"relationship_effective_date" bson:"relationship_effective_date"`
	Student                   *bool      `json:"student" bson:"student"`
	OverAgeDisabled           *bool      `json:"over_age_disabled" bson:"over_age_disabled"`
	TobaccoUser               *bool      `json:"tobacco_user" bson:"tobacco_user"`
	DependentId               string     `json:"dependent_id" bson:"dependent_id"`
}

type Dependents []Dependent

type EmployeeDTO struct {
	// User profile fields
	EmployeeId                   string     `json:"employee_id" bson:"employee_id"`
	FirstName                    string     `json:"first_name" bson:"first_name"`
	LastName                     string     `json:"last_name" bson:"last_name"`
	Email                        string     `json:"email" bson:"email"`
	PreferredFirstName           string     `json:"preferred_first_name" bson:"preferred_first_name"`
	Sex                          string     `json:"sex" bson:"sex"`
	DateOfBirth                  *time.Time `json:"date_of_birth" bson:"date_of_birth"`
	PhoneNumber                  string     `json:"phone_number" bson:"phone_number"`
	Locale                       string     `json:"locale" bson:"locale"`
	NationalIdentificationNumber string     `json:"national_identification_number" bson:"national_identification_number"`
	RegisteredIndianAct          *bool      `json:"registered_indian_act" bson:"registered_indian_act"`
	TobaccoUser                  *bool      `json:"tobacco_user" bson:"tobacco_user"`
	Address1                     string     `json:"address1" bson:"address1"`
	Address2                     string     `json:"address2" bson:"address2"`
	City                         string     `json:"city" bson:"city"`
	Province                     string     `json:"province" bson:"province"`
	Country                      string     `json:"country" bson:"country"`
	PostalCode                   string     `json:"postal_code" bson:"postal_code"`

	// Group membership fields
	GroupId                         string     `json:"group_id" bson:"group_id"`
	BenefitClass                    string     `json:"benefit_class" bson:"benefit_class"`
	DateOfHire                      *time.Time `json:"date_of_hire" bson:"date_of_hire"`
	BenefitsStartDate               *time.Time `json:"benefits_start_date" bson:"benefits_start_date"`
	ProvinceOfEmployment            string     `json:"province_of_employment" bson:"province_of_employment"`
	AnnualEarnings                  float64    `json:"annual_earnings" bson:"annual_earnings"`
	AnnualEarningsForPooledBenefits float64    `json:"annual_earnings_for_pooled_benefits" bson:"annual_earnings_for_pooled_benefits"`
	AnnualEarningsEffectiveDate     *time.Time `json:"annual_earnings_effective_date" bson:"annual_earnings_effective_date"`
	HrsWorkedPerWeek                float64    `json:"hrs_worked_per_week" bson:"hrs_worked_per_week"`
	Title                           string     `json:"title" bson:"title"`
	OfficeLocation                  string     `json:"office_location" bson:"office_location"`
	EmploymentStatus                string     `json:"employment_status" bson:"employment_status"`
	Occupation                      string     `json:"occupation" bson:"occupation"`
	ActivationDate                  *time.Time `json:"activation_date" bson:"activation_date"`
	SuspensionType                  string     `json:"suspension_type" bson:"suspension_type"`
	SuspensionReason                string     `json:"suspension_reason" bson:"suspension_reason"`
	SuspendedDate                   *time.Time `json:"suspended_date" bson:"suspended_date"`
	BillingDivision                 string     `json:"billing_division" bson:"billing_division"`
	PayGroup                        string     `json:"pay_group" bson:"pay_group"`
	EmployeeLeave                   string     `json:"employee_leave" bson:"employee_leave"`
	EmployeeLeaveStartDate          *time.Time `json:"employee_leave_start_date" bson:"employee_leave_start_date"`
	EnrollmentEndDate               *time.Time `json:"enrollment_end_date" bson:"enrollment_end_date"`
	NoPlatformFeeCharges            bool       `json:"no_platform_fee_charges" bson:"no_platform_fee_charges"`
	Department                      string     `json:"department" bson:"department"`
	CustomFields                    msg.M      `json:"custom_fields" bson:"custom_fields"`
	BenefitClassChangeEffectiveDate time.Time  `json:"benefit_class_change_effective_date" bson:"benefit_class_change_effective_date"`

	// dependents
	Dependents Dependents `json:"dependents" bson:"dependents"`

	// benefits
	Benefits Benefits `json:"benefits" bson:"benefits"`

	// control fields
	State              State     `json:"state" bson:"state"`
	StateEffectiveDate time.Time `json:"state_effective_date" bson:"state_effective_date"`
	FieldsToUpdate     []string  `json:"fields_to_update" bson:"fields_to_update"`

	// Special rules
	OnBenefitClassChange string `json:"on_benefit_class_change" bson:"on_benefit_class_change"`
	OnReinstate          string `json:"on_reinstate" bson:"on_reinstate"`
}
