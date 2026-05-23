# Debtease — Debt Obligation Architecture & Schema Evolution

## Technical Design Document (TDD) — Go Backend v2

---

# 1. Document Overview

---

## Purpose

This document defines the evolution of the Debtease debt system from:

```text id="xy04wz"
loan-centric architecture
```

to:

```text id="ylgf8d"
financial obligation architecture
```

The new architecture enables deterministic support for:

* secured loans
* unsecured loans
* BNPL
* 0% EMI
* revolving debt
* consumer finance
* hybrid repayment products

while remaining aligned with:

* Debt Classification Domain
* Stress Simulation Domain
* Financial Snapshot Architecture
* Financial Health Score System

---

## Primary Goals

The system must:

1. Normalize all debt instruments into a canonical structure
2. Remove loan-type assumptions from financial engines
3. Support future debt products without schema rewrites
4. Preserve deterministic recomputation
5. Maintain explainability and auditability

---

# 2. Existing System Problems

---

# 2.1 Current Architecture Assumption

Current schema assumes:

```text id="ijbxpf"
Debt == amortized EMI loan
```

This assumption leaks into:

* schema design
* calculations
* stress simulation
* onboarding inputs
* health scoring

---

# 2.2 Why This Fails

The following products do not behave like traditional loans:

| Product          | Breaks Current Model         |
| ---------------- | ---------------------------- |
| BNPL             | No traditional EMI structure |
| Credit Card      | Revolving repayment          |
| 0% EMI           | Interest-free but rigid      |
| LOC              | Dynamic utilization          |
| Salary Advance   | Ultra short-term             |
| Merchant Finance | Flexible tenure              |

---

# 2.3 Architectural Consequences

Without normalization:

* debt-type branching spreads everywhere
* stress logic becomes inconsistent
* calculations become non-deterministic
* schema becomes unmaintainable
* future products become expensive to add

---

# 3. New Architectural Principle

---

# 3.1 Core Principle

Debtease does not reason about:

```text id="ayv4ce"
loan products
```

It reasons about:

```text id="ngb1ch"
financial obligations
```

---

# 3.2 Canonical Flow

```text id="mxzvwv"
User Debt Input
        ↓
Debt Instrument Classification
        ↓
Normalization Layer
        ↓
Canonical Obligation Model
        ↓
Financial Engines
```

---

# 3.3 Engine Isolation Principle

All downstream systems consume only:

```go id="uw9f8j"
NormalizedDebt
```

No engine may branch directly on:

* loan type
* UI labels
* frontend logic

---

# 4. Domain Model

---

# 4.1 Debt Instrument Types

---

## Purpose

Represents:

* real-world debt products
* user-selectable debt categories

NOT repayment behavior.

---

## PostgreSQL Enum

```sql id="v6hczf"
CREATE TYPE debt_instrument_type AS ENUM (
    'home_loan',
    'car_loan',
    'gold_loan',

    'personal_loan',
    'education_loan',
    'medical_loan',

    'consumer_emi',

    'bnpl',
    'salary_advance',
    'pay_later',

    'credit_card',
    'line_of_credit'
);
```

---

## Go Enum

```go id="j9l3u7"
type DebtInstrumentType string
```

---

# 4.2 Interest Models

---

## Purpose

Represents repayment and interest behavior independently from debt type.

Critical for:

* 0% EMI
* revolving debt
* floating-rate products

---

## PostgreSQL Enum

```sql id="2jwkgi"
CREATE TYPE interest_model AS ENUM (
    'fixed',
    'floating',
    'zero_cost_emi',
    'revolving'
);
```

---

## Examples

| Instrument    | Interest Model |
| ------------- | -------------- |
| Home Loan     | floating       |
| Personal Loan | fixed          |
| Consumer EMI  | zero_cost_emi  |
| Credit Card   | revolving      |

---

# 4.3 Structural Characteristics

Derived deterministic metadata.

---

## Purpose

Financial engines require:

* structural semantics
* not raw debt labels

---

## Go Structure

```go id="chvq12"
type DebtCharacteristics struct {
    Secured bool

    Revolving bool
    ShortTerm bool

    FixedObligation bool
    InterestBearing bool

    FragmentationWeight float64
    StressPriority int
}
```

---

# 4.4 Canonical Debt Object

---

## Purpose

Single normalized representation consumed by all engines.

---

## Go Structure

```go id="l4r3c2"
type NormalizedDebt struct {
    DebtID uuid.UUID

    InstrumentType DebtInstrumentType
    InterestModel InterestModel

    OutstandingBalance decimal.Decimal

    MonthlyObligation decimal.Decimal
    MinimumObligation decimal.Decimal

    RemainingMonths int

    Characteristics DebtCharacteristics

    ClassificationVersion string
    PolicyVersion string
}
```

---

# 5. Database Schema Evolution

---

# 5.1 Existing Debts Table Problems

Current fields tightly couple schema to EMI loans:

| Existing Field | Problem                          |
| -------------- | -------------------------------- |
| type           | Too generic                      |
| emi_amount     | EMI-specific                     |
| min_payment    | Credit-card specific             |
| tenure_months  | Original loan-centric            |
| months_paid    | Derived state                    |
| risk_level     | Violates architecture philosophy |

---

# 5.2 Updated Debts Table

---

## Final Schema

```sql id="m7t6j2"
CREATE TABLE debts (
    id UUID PRIMARY KEY,

    user_id UUID NOT NULL REFERENCES users(id),

    name TEXT NOT NULL,
    lender TEXT,

    instrument_type debt_instrument_type NOT NULL,

    interest_model interest_model NOT NULL,

    original_principal NUMERIC(14,2),
    outstanding_balance NUMERIC(14,2) NOT NULL,

    monthly_obligation NUMERIC(14,2) NOT NULL,
    minimum_obligation NUMERIC(14,2),

    remaining_months INTEGER,

    due_day INTEGER,

    is_secured BOOLEAN NOT NULL,
    is_revolving BOOLEAN NOT NULL,
    is_short_term BOOLEAN NOT NULL,
    is_interest_bearing BOOLEAN NOT NULL,

    fragmentation_weight NUMERIC(4,2) NOT NULL,
    stress_priority SMALLINT NOT NULL,

    treatment_classification TEXT,

    engine_version TEXT NOT NULL,
    policy_version TEXT NOT NULL,

    extra_metadata JSONB,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

# 5.3 Why Monthly Obligation Replaces EMI

---

## Previous

```text id="4umtbz"
emi_amount
```

---

## Problem

Not all obligations are EMIs:

* BNPL dues
* minimum credit card dues
* LOC payments
* merchant financing

---

## New Model

```text id="87fjp5"
monthly_obligation
```

Unified abstraction.

---

# 5.4 Why Remaining Months Replaces Tenure

---

## Previous

```text id="ly2jhy"
tenure_months
```

---

## Problem

Original tenure is not useful for:

* stress
* rigidity
* survivability

---

## New

```text id="d4yzrs"
remaining_months
```

Tracks:

* remaining lock-in
* remaining rigidity

Aligned with:

* Rigidity Score computation 

---

# 5.5 Removal of Risk Level

---

## Existing

```text id="fpk0d6"
risk_level
```

---

## Problem

Violates domain philosophy.

Debtease does NOT:

* predict defaults
* assign borrower risk

Aligned with domain constraints. 

---

## Replacement

```text id="07xz9g"
treatment_classification
```

Values:

* SAFE_TO_HOLD
* NEEDS_MONITORING
* AVOID_ADDING
* ACTIVELY_REDUCE

---

# 6. Credit Card Metadata Table

---

# 6.1 Purpose

Credit cards require:

* statement cycles
* billing periods
* grace logic

without polluting normalized debt schema.

---

# 6.2 Schema

```sql id="z3hx4r"
CREATE TABLE debt_credit_card_metadata (
    debt_id UUID PRIMARY KEY REFERENCES debts(id),

    billing_day SMALLINT NOT NULL,
    grace_period_days SMALLINT NOT NULL,

    last_statement_date DATE,
    payment_due_date DATE,

    billed_balance NUMERIC(14,2),
    unbilled_balance NUMERIC(14,2),

    minimum_due NUMERIC(14,2),

    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
```

---

# 7. Debt Normalization Layer

---

# 7.1 Purpose

Normalization isolates:

* debt complexity
* product-specific semantics
* repayment differences

from financial engines.

---

# 7.2 Responsibilities

The normalization layer must:

1. Validate raw input
2. Derive characteristics
3. Normalize obligations
4. Apply policy defaults
5. Generate canonical debt object

---

# 7.3 Package Structure

```text id="3tzhqy"
internal/domain/debt/
├── types.go
├── classifier.go
├── normalizer.go
├── validator.go
├── mapper.go
├── policy.go
├── repository.go
├── service.go
└── test/
```

---

# 7.4 Main Interface

```go id="4jew4o"
type Normalizer interface {
    Normalize(input DebtInput) (*NormalizedDebt, error)
}
```

---

# 8. Debt Classification Rules

---

# 8.1 Purpose

Maps debt instrument types into structural semantics.

---

# 8.2 Example — BNPL

```go id="2d1t9d"
func bnplCharacteristics() DebtCharacteristics {
    return DebtCharacteristics{
        Secured: false,

        Revolving: false,
        ShortTerm: true,

        FixedObligation: true,
        InterestBearing: false,

        FragmentationWeight: 0.5,
        StressPriority: 1,
    }
}
```

Aligned with:

* fragmentation model 
* stress failure ordering 

---

# 8.3 Example — Home Loan

```go id="fq1c0o"
func homeLoanCharacteristics() DebtCharacteristics {
    return DebtCharacteristics{
        Secured: true,

        Revolving: false,
        ShortTerm: false,

        FixedObligation: true,
        InterestBearing: true,

        FragmentationWeight: 1.0,
        StressPriority: 4,
    }
}
```

---

# 9. BNPL Handling

---

# 9.1 Important Principle

BNPL is:

* a debt instrument
* not a loan subtype

---

# 9.2 Input Example

```json id="jm9i4m"
{
  "instrument_type": "bnpl",
  "monthly_due": 2500,
  "remaining_months": 3
}
```

---

# 9.3 Normalized Result

```json id="om9p8q"
{
  "monthly_obligation": 2500,
  "remaining_months": 3,
  "short_term": true,
  "fragmentation_weight": 0.5
}
```

---

# 10. 0% EMI Handling

---

# 10.1 Critical Principle

0% EMI is NOT:

* a debt type
* a product category

It is:

* an interest model

---

# 10.2 Example

```json id="vyj76t"
{
  "instrument_type": "consumer_emi",
  "interest_model": "zero_cost_emi"
}
```

---

# 10.3 Why This Matters

Even with zero interest:

* rigidity exists
* fragmentation exists
* stress impact exists

Only interest-rate shock differs.

Aligned with stress simulation rules. 

---

# 11. Financial Snapshot Evolution

---

# 11.1 Purpose

Improve snapshot performance and aggregation efficiency.

---

# 11.2 New Aggregate Fields

```sql id="u9sfsr"
ALTER TABLE financial_snapshots
ADD COLUMN unsecured_monthly_obligation NUMERIC(14,2),

ADD COLUMN revolving_monthly_obligation NUMERIC(14,2),

ADD COLUMN short_term_obligations NUMERIC(14,2);
```

---

# 11.3 Why

Classification and stress repeatedly compute:

* unsecured exposure
* short-term rigidity
* fragmentation

Pre-aggregating improves:

* dashboard latency
* recalculation efficiency
* stress simulation performance

---

# 12. Engine Integration

---

# 12.1 Classification Engine

Consumes:

```go id="tw9g8v"
[]NormalizedDebt
```

Calculates:

* EMI ratio
* fragmentation
* rigidity
* unsecured exposure

Aligned with classification domain metrics. 

---

# 12.2 Stress Engine

Consumes:

* normalized obligations
* stress priorities
* rigidity metadata

Aligned with:

* deterministic shock modeling
* failure ordering 

---

# 12.3 Financial Health Score

Uses:

* obligation totals
* fragmentation weights
* unsecured ratios

Aligned with health scoring system. 

---

# 13. API Contract

---

# POST /debts

---

## Request

```json id="9gbsh9"
{
  "instrument_type": "bnpl",
  "monthly_due": 3200,
  "remaining_months": 4
}
```

---

## Response

```json id="x3dktm"
{
  "debt_id": "uuid",

  "normalized": {
    "monthly_obligation": 3200,
    "short_term": true,
    "secured": false
  }
}
```

---

# 14. Validation Rules

---

# 14.1 BNPL

Required:

* monthly_due
* remaining_months

Forbidden:

* long tenure (>24 months)

---

# 14.2 Credit Card

Required:

* outstanding_balance

Optional:

* minimum_obligation

---

# 14.3 Home Loan

Required:

* monthly_obligation
* remaining_months

---

# 15. Migration Strategy

---

# Phase 1 — Schema Evolution

Add:

* instrument_type
* interest_model
* monthly_obligation
* structural flags

Keep old fields temporarily.

---

# Phase 2 — Normalization Layer

Introduce:

* classifier
* validator
* normalized debt generation

---

# Phase 3 — Engine Migration

Switch:

* classification
* stress
* health scoring

to:

* normalized obligations

---

# Phase 4 — Remove Legacy Fields

Drop:

* emi_amount
* tenure_months
* risk_level
* months_paid

---

# 16. Testing Strategy

---

# 16.1 Golden Tests

Mandatory.

Guarantee:

* deterministic normalization
* stable outputs

---

# 16.2 Validation Tests

Test:

* BNPL edge cases
* revolving debt
* 0% EMI
* missing obligations

---

# 16.3 Stress Ordering Tests

Must validate:

```text id="ng7z53"
BNPL
→ short-term unsecured
→ long-term unsecured
→ secured
```

Aligned with stress simulation policy. 

---

# 17. Non-Goals

Explicitly excluded from v2:

* probabilistic risk scoring
* underwriting
* repayment optimization
* ML classification
* behavioral predictions

---

# 18. Critical Invariants

Must always hold:

1. Same input → same normalized debt
2. Engines consume normalized obligations only
3. Interest model independent from debt type
4. Structural metadata deterministic
5. Frontend never determines risk semantics
6. Hypothetical loans never mutate real debt state

---

# 19. Final Architecture

```text id="e0rty2"
User Input
    ↓
Debt Instrument Classification
    ↓
Normalization Layer
    ↓
Canonical Obligation Model
    ↓
Classification Engine
    ↓
Stress Simulation
    ↓
Financial Health Engine
```

---

# 20. Success Criteria

Implementation succeeds when:

* BNPL works without hacks
* 0% EMI works without branching
* credit cards integrate cleanly
* engines remain deterministic
* future debt products require only classifier additions
* schema remains stable across new instruments
* stress simulation remains explainable
* financial health remains recomputable
