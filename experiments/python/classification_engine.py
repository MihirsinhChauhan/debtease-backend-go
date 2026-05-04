from __future__ import annotations

from dataclasses import dataclass
from decimal import Decimal, getcontext
from typing import Dict, List, Optional

from loan_portfolio_engine import EMIRecord, Loan


# Use higher precision for ratio calculations
getcontext().prec = 28


# ----------------------------
# Domain types (ported from Go)
# ----------------------------

TreatmentCategory = str

SAFE_TO_HOLD: TreatmentCategory = "SAFE_TO_HOLD"
NEEDS_MONITORING: TreatmentCategory = "NEEDS_MONITORING"
AVOID_ADDING: TreatmentCategory = "AVOID_ADDING"
ACTIVELY_REDUCE: TreatmentCategory = "ACTIVELY_REDUCE"


Flag = str

FLAG_HIGH_CASHFLOW_RIGIDITY: Flag = "HIGH_CASHFLOW_RIGIDITY"
FLAG_SHORT_TERM_DEBT_STACKING: Flag = "SHORT_TERM_DEBT_STACKING"
FLAG_LOW_SHOCK_BUFFER: Flag = "LOW_SHOCK_BUFFER"
FLAG_UNSECURED_OVEREXPOSURE: Flag = "UNSECURED_OVEREXPOSURE"


ConfidenceLevel = str

CONFIDENCE_LOW: ConfidenceLevel = "LOW"
CONFIDENCE_MEDIUM: ConfidenceLevel = "MEDIUM"
CONFIDENCE_HIGH: ConfidenceLevel = "HIGH"


DebtInstrumentType = str

INSTRUMENT_BNPL: DebtInstrumentType = "bnpl"
INSTRUMENT_LOC: DebtInstrumentType = "loc"
INSTRUMENT_CREDIT_CARD: DebtInstrumentType = "credit_card"
INSTRUMENT_PERSONAL_LOAN: DebtInstrumentType = "personal_loan"
INSTRUMENT_STUDENT_LOAN: DebtInstrumentType = "student_loan"
INSTRUMENT_MORTGAGE: DebtInstrumentType = "mortgage"
INSTRUMENT_MEDICAL: DebtInstrumentType = "medical"
INSTRUMENT_OTHER: DebtInstrumentType = "other"


decimal_zero = Decimal("0")
half = Decimal("0.5")


def ratio(num: Decimal, denom: Decimal) -> Decimal:
    """Safe division: returns 0 when denom is zero."""
    if denom == decimal_zero:
        return decimal_zero
    return num / denom


@dataclass
class DebtEntry:
    id: str
    emi: Decimal
    remaining_months: int
    secured: bool
    instrument_type: DebtInstrumentType


@dataclass
class FinancialContext:
    monthly_net_income: Decimal
    fixed_obligations: Decimal
    existing_emis: Decimal
    short_term_obligations_count: int


@dataclass
class ClassificationInput:
    debts: List[DebtEntry]
    context: FinancialContext


@dataclass
class ContextMetrics:
    emi_ratio: Decimal
    fragmentation_index: Decimal
    rigidity_score: Decimal
    shock_buffer_months: Decimal
    unsecured_ratio: Decimal


@dataclass
class ClassificationResult:
    treatment: TreatmentCategory
    flags: List[Flag]
    confidence: ConfidenceLevel
    metrics: ContextMetrics
    engine_version: str


@dataclass
class PolicyThresholds:
    max_safe_emi_ratio: Decimal       # e.g. 0.35
    max_fragmentation_index: Decimal  # e.g. 4
    min_shock_buffer_months: Decimal  # e.g. 1.5
    max_unsecured_ratio: Decimal      # e.g. 0.6


def default_policy_thresholds() -> PolicyThresholds:
    """India v1 defaults (mirrors Go defaults)."""
    return PolicyThresholds(
        max_safe_emi_ratio=Decimal("0.35"),
        max_fragmentation_index=Decimal("4"),
        min_shock_buffer_months=Decimal("1.5"),
        max_unsecured_ratio=Decimal("0.60"),
    )


ENGINE_VERSION = "v1.0.0"


# ----------------------------
# Metric computations
# ----------------------------

def _total_emi(debts: List[DebtEntry]) -> Decimal:
    total = decimal_zero
    for d in debts:
        total += d.emi
    return total


def _count_bnpl(debts: List[DebtEntry]) -> int:
    return sum(1 for d in debts if d.instrument_type == INSTRUMENT_BNPL)


def compute_emi_ratio(input: ClassificationInput) -> Decimal:
    total_emi = _total_emi(input.debts)
    return ratio(total_emi, input.context.monthly_net_income)


def compute_fragmentation_index(input: ClassificationInput) -> Decimal:
    n = len(input.debts)
    if n == 0:
        return decimal_zero
    active_debts = Decimal(n)
    bnpl_count = Decimal(_count_bnpl(input.debts))
    return active_debts + bnpl_count * half


def compute_rigidity_score(input: ClassificationInput) -> Decimal:
    total = decimal_zero
    for d in input.debts:
        rem = Decimal(d.remaining_months)
        total += d.emi * rem
    return ratio(total, input.context.monthly_net_income)


def compute_shock_buffer_months(input: ClassificationInput) -> Decimal:
    total_emi = _total_emi(input.debts)
    numerator = (
        input.context.monthly_net_income
        - input.context.fixed_obligations
        - total_emi
    )
    return ratio(numerator, total_emi)


def compute_unsecured_ratio(input: ClassificationInput) -> Decimal:
    total_emi = _total_emi(input.debts)
    if total_emi == decimal_zero:
        return decimal_zero
    unsecured = decimal_zero
    for d in input.debts:
        if not d.secured:
            unsecured += d.emi
    return ratio(unsecured, total_emi)


def compute_all_metrics(input: ClassificationInput) -> ContextMetrics:
    return ContextMetrics(
        emi_ratio=compute_emi_ratio(input),
        fragmentation_index=compute_fragmentation_index(input),
        rigidity_score=compute_rigidity_score(input),
        shock_buffer_months=compute_shock_buffer_months(input),
        unsecured_ratio=compute_unsecured_ratio(input),
    )


# ----------------------------
# Classifier
# ----------------------------

def _contains_flag(flags: List[Flag], flag: Flag) -> bool:
    return flag in flags


def _apply_thresholds(
    input: ClassificationInput, metrics: ContextMetrics, t: PolicyThresholds
) -> List[Flag]:
    flags: List[Flag] = []

    if metrics.emi_ratio > t.max_safe_emi_ratio:
        flags.append(FLAG_HIGH_CASHFLOW_RIGIDITY)

    if metrics.fragmentation_index >= t.max_fragmentation_index:
        flags.append(FLAG_SHORT_TERM_DEBT_STACKING)

    # Only flag low shock buffer when there are debts (otherwise buffer is not meaningful).
    if input.debts and metrics.shock_buffer_months < t.min_shock_buffer_months:
        flags.append(FLAG_LOW_SHOCK_BUFFER)

    if (
        metrics.unsecured_ratio > t.max_unsecured_ratio
        and metrics.unsecured_ratio != decimal_zero
    ):
        flags.append(FLAG_UNSECURED_OVEREXPOSURE)

    return flags


def _assign_treatment(
    flags: List[Flag], metrics: ContextMetrics, t: PolicyThresholds
) -> TreatmentCategory:
    has_high_rigidity = _contains_flag(flags, FLAG_HIGH_CASHFLOW_RIGIDITY)
    has_stacking = _contains_flag(flags, FLAG_SHORT_TERM_DEBT_STACKING)
    has_low_buffer = _contains_flag(flags, FLAG_LOW_SHOCK_BUFFER)
    has_unsecured = _contains_flag(flags, FLAG_UNSECURED_OVEREXPOSURE)

    # ACTIVELY_REDUCE: high unsecured + (high rigidity or low buffer)
    if has_unsecured and (has_high_rigidity or has_low_buffer):
        return ACTIVELY_REDUCE

    # AVOID_ADDING: high fragmentation OR low buffer
    if has_stacking or has_low_buffer:
        return AVOID_ADDING

    # NEEDS_MONITORING: moderate stress (e.g. high EMI ratio but structure stable)
    if has_high_rigidity:
        return NEEDS_MONITORING

    # SAFE_TO_HOLD: no flags or only unsecured without other stress
    return SAFE_TO_HOLD


def _compute_confidence(input: ClassificationInput, flags: List[Flag]) -> ConfidenceLevel:
    if input.context.monthly_net_income <= decimal_zero:
        return CONFIDENCE_LOW
    if not input.debts:
        return CONFIDENCE_LOW
    if len(flags) >= 2:
        return CONFIDENCE_HIGH
    if len(flags) == 1:
        return CONFIDENCE_MEDIUM
    return CONFIDENCE_HIGH


def classify_with_policy(
    input: ClassificationInput, thresholds: PolicyThresholds
) -> ClassificationResult:
    metrics = compute_all_metrics(input)
    flags = _apply_thresholds(input, metrics, thresholds)
    treatment = _assign_treatment(flags, metrics, thresholds)
    confidence = _compute_confidence(input, flags)
    return ClassificationResult(
        treatment=treatment,
        flags=flags,
        confidence=confidence,
        metrics=metrics,
        engine_version=ENGINE_VERSION,
    )


def classify(input: ClassificationInput) -> ClassificationResult:
    return classify_with_policy(input, default_policy_thresholds())


# ----------------------------
# Adapters from repayment engine
# ----------------------------

def _last_month_by_loan(schedule: Dict[int, List[EMIRecord]]) -> Dict[str, int]:
    """
    Returns the final month index for each loan in the schedule.
    We assume schedule contains entries until each loan is fully repaid.
    """
    last: Dict[str, int] = {}
    for month, records in schedule.items():
        for r in records:
            current = last.get(r.loan_id)
            if current is None or month > current:
                last[r.loan_id] = month
    return last


def schedule_to_classification_input(
    schedule: Dict[int, List[EMIRecord]],
    loans: List[Loan],
    month: int,
    monthly_net_income: float,
    fixed_obligations: float,
    extra_short_term_obligations: int = 0,
) -> ClassificationInput:
    """
    Build ClassificationInput from a repayment schedule snapshot at a given month.
    Monetary values are converted to Decimal at the boundary.
    """
    last_month = _last_month_by_loan(schedule)

    # Index loans by id to access instrument_type / secured flags.
    loan_index: Dict[str, Loan] = {loan.loan_id: loan for loan in loans}

    debts: List[DebtEntry] = []
    total_emi_decimal = decimal_zero
    short_term_count = extra_short_term_obligations

    records = schedule.get(month, [])
    for r in records:
        loan = loan_index.get(r.loan_id)
        if loan is None:
            continue

        remaining_months = max(0, last_month.get(r.loan_id, month) - month + 1)

        emi_dec = Decimal(str(r.emi))
        total_emi_decimal += emi_dec

        instrument_type: DebtInstrumentType = getattr(
            loan, "instrument_type", INSTRUMENT_OTHER
        )
        secured: bool = bool(getattr(loan, "secured", False))

        debts.append(
            DebtEntry(
                id=r.loan_id,
                emi=emi_dec,
                remaining_months=remaining_months,
                secured=secured,
                instrument_type=instrument_type,
            )
        )

        # Short-term obligations count: BNPL + LOC in v1.
        if instrument_type in (INSTRUMENT_BNPL, INSTRUMENT_LOC):
            short_term_count += 1

    context = FinancialContext(
        monthly_net_income=Decimal(str(monthly_net_income)),
        fixed_obligations=Decimal(str(fixed_obligations)),
        existing_emis=total_emi_decimal,
        short_term_obligations_count=short_term_count,
    )

    return ClassificationInput(debts=debts, context=context)


__all__ = [
    # types
    "TreatmentCategory",
    "SAFE_TO_HOLD",
    "NEEDS_MONITORING",
    "AVOID_ADDING",
    "ACTIVELY_REDUCE",
    "Flag",
    "FLAG_HIGH_CASHFLOW_RIGIDITY",
    "FLAG_SHORT_TERM_DEBT_STACKING",
    "FLAG_LOW_SHOCK_BUFFER",
    "FLAG_UNSECURED_OVEREXPOSURE",
    "ConfidenceLevel",
    "CONFIDENCE_LOW",
    "CONFIDENCE_MEDIUM",
    "CONFIDENCE_HIGH",
    "DebtInstrumentType",
    "INSTRUMENT_BNPL",
    "INSTRUMENT_LOC",
    "INSTRUMENT_CREDIT_CARD",
    "INSTRUMENT_PERSONAL_LOAN",
    "INSTRUMENT_STUDENT_LOAN",
    "INSTRUMENT_MORTGAGE",
    "INSTRUMENT_MEDICAL",
    "INSTRUMENT_OTHER",
    "DebtEntry",
    "FinancialContext",
    "ClassificationInput",
    "ContextMetrics",
    "ClassificationResult",
    "PolicyThresholds",
    # engine
    "default_policy_thresholds",
    "classify",
    "classify_with_policy",
    "compute_all_metrics",
    # adapters
    "schedule_to_classification_input",
]

