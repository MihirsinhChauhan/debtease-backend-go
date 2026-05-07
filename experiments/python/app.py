import streamlit as st
import pandas as pd
from io import BytesIO

from loan_portfolio_engine import Loan
from strategy_engine_v2 import simulate_with_strategy, StrategyType
from metrics_service import (
    summarize_portfolio,
    debt_free_month,
    first_emi_relief_month,
    principal_delta_at,
    emi_drop_events,
)
from classification_engine import (
    classify,
    schedule_to_classification_input,
    FLAG_HIGH_CASHFLOW_RIGIDITY,
    FLAG_SHORT_TERM_DEBT_STACKING,
    FLAG_LOW_SHOCK_BUFFER,
    FLAG_UNSECURED_OVEREXPOSURE,
)
from stress_engine import simulate_stress

try:
    from reportlab.lib.pagesizes import A4
    from reportlab.pdfgen import canvas
    REPORTLAB_AVAILABLE = True
except ImportError:
    REPORTLAB_AVAILABLE = False

st.set_page_config(layout="wide")
st.title("DebtEase — Product Simulator")

# =====================
# 1. SIDEBAR — User Financial State
# =====================
st.sidebar.header("User Financial State")

monthly_income = st.sidebar.number_input(
    "Monthly income",
    min_value=0.0,
    value=60_000.0,
    step=1000.0,
)
fixed_expenses = st.sidebar.number_input(
    "Fixed expenses",
    min_value=0.0,
    value=10_000.0,
    step=1000.0,
)

st.sidebar.header("Loans")

loan_count = st.sidebar.number_input("Number of loans", 1, 10, 2)

loans = []
for i in range(loan_count):
    st.sidebar.subheader(f"Loan {i + 1}")
    loan_id = st.sidebar.text_input(f"Loan ID {i + 1}", f"L{i + 1}", key=f"lid_{i}")
    principal = st.sidebar.number_input(
        f"Principal {i + 1}",
        min_value=1_000.0,
        value=5_00_000.0,
        step=10_000.0,
        key=f"principal_{i}",
    )
    rate = st.sidebar.number_input(
        f"Interest % {i + 1}",
        min_value=0.0,
        value=10.0,
        step=0.5,
        key=f"rate_{i}",
    )
    tenure = st.sidebar.number_input(
        f"Tenure (months) {i + 1}",
        min_value=1,
        value=60,
        key=f"tenure_{i}",
    )
    instrument_type = st.sidebar.selectbox(
        f"Type {i + 1}",
        options=["personal_loan", "vehicle_loan", "mortgage", "consumer_durable", "bnpl", "loc", "other"],
        index=0,
        format_func=lambda x: x.replace("_", " ").title(),
        key=f"instrument_{i}",
    )
    secured = st.sidebar.checkbox(
        f"Secured {i + 1}",
        value=(instrument_type == "mortgage"),
        key=f"secured_{i}",
    )

    loans.append(
        Loan(
            loan_id=loan_id,
            principal=principal,
            annual_rate=rate,
            tenure_months=tenure,
            instrument_type=instrument_type,
            secured=secured,
        )
    )

extra_bnpl_count = st.sidebar.number_input(
    "BNPL count (outside listed loans)",
    min_value=0,
    max_value=50,
    value=0,
)

st.sidebar.header("Strategy")

strategy = st.sidebar.selectbox(
    "Prepayment strategy",
    [StrategyType.SNOWBALL, StrategyType.AVALANCHE],
)
monthly_extra = st.sidebar.number_input(
    "Monthly extra prepayment",
    min_value=0.0,
    value=5_000.0,
    step=1000.0,
)

# =====================
# 2. ENGINE LAYER — run pipeline, store in state
# =====================
analysis_month = 1

baseline_schedule = simulate_with_strategy(loans, strategy=StrategyType.BASELINE)
strategy_schedule = simulate_with_strategy(
    loans,
    strategy=strategy,
    monthly_extra=monthly_extra,
)

base_summary = summarize_portfolio(baseline_schedule)
strat_summary = summarize_portfolio(strategy_schedule)

classification_input = schedule_to_classification_input(
    strategy_schedule,
    loans,
    analysis_month,
    monthly_net_income=monthly_income,
    fixed_obligations=fixed_expenses,
    extra_short_term_obligations=extra_bnpl_count,
)
classification_result = classify(classification_input)

stress_result = simulate_stress(
    income=monthly_income,
    fixed_expenses=fixed_expenses,
    schedule=strategy_schedule,
    loans=loans,
    month=analysis_month,
)

# Current flexibility buffer (no shock): disposable - total EMI
total_emi_month1 = sum(r.emi for r in strategy_schedule.get(analysis_month, []))
flexibility_buffer_current = monthly_income - fixed_expenses - total_emi_month1

debt_free = debt_free_month(strategy_schedule)
baseline_debt_free = debt_free_month(baseline_schedule)
baseline_relief = first_emi_relief_month(baseline_schedule)
strategy_relief = first_emi_relief_month(strategy_schedule)
interest_saved = base_summary["total_interest"] - strat_summary["total_interest"]

state = {
    "loan_metrics": {
        "total_emi": total_emi_month1,
        "debt_free_month": debt_free,
        "baseline_debt_free_month": baseline_debt_free,
        "flexibility_buffer": flexibility_buffer_current,
        "interest_saved": interest_saved,
        "baseline_summary": base_summary,
        "strategy_summary": strat_summary,
    },
    "classification": {
        "treatment": classification_result.treatment,
        "flags": classification_result.flags,
        "confidence": classification_result.confidence,
        "metrics": {
            "emi_ratio": float(classification_result.metrics.emi_ratio),
            "fragmentation_index": float(classification_result.metrics.fragmentation_index),
            "rigidity_score": float(classification_result.metrics.rigidity_score),
            "shock_buffer_months": float(classification_result.metrics.shock_buffer_months),
            "unsecured_ratio": float(classification_result.metrics.unsecured_ratio),
        },
    },
    "stress": {
        "stress_level": stress_result.stress_level,
        "flexibility_after_shock": stress_result.flexibility_after_shock,
        "survivability_band": stress_result.survivability_band,
        "failed_debt_categories": stress_result.failed_debt_categories,
        "key_drivers": stress_result.key_drivers,
    },
    "repayment": {
        "baseline_schedule": baseline_schedule,
        "strategy_schedule": strategy_schedule,
    },
    "prepayment": {
        "monthly_extra": monthly_extra,
        "strategy": strategy,
    },
}

# Human-readable debt health label
def debt_health_label(level: str) -> str:
    return {"RESILIENT": "Resilient", "SENSITIVE": "Sensitive", "FRAGILE": "Fragile"}.get(level, level)

# Flag to short explanation for Debt Health tab
def flag_to_bullet(flag: str) -> str:
    m = {
        FLAG_HIGH_CASHFLOW_RIGIDITY: "Cashflow rigidity",
        FLAG_SHORT_TERM_DEBT_STACKING: "Short-term debt stacking",
        FLAG_LOW_SHOCK_BUFFER: "Low shock buffer",
        FLAG_UNSECURED_OVEREXPOSURE: "High unsecured exposure",
    }
    return m.get(flag, flag.replace("_", " ").lower())

# Snapshot explanation (one line, no raw metrics)
def snapshot_explanation(stress_level: str, flags: list) -> str:
    if stress_level == "FRAGILE":
        return "Your current loans leave very limited room for income disruption."
    if stress_level == "SENSITIVE":
        return "Your current loans leave limited room for income disruption."
    if flags:
        return "Your structure is manageable; watch the factors below."
    return "Your debt structure looks manageable with room for flexibility."

# Strategy explanation from treatment
def strategy_explanation(treatment: str) -> str:
    m = {
        "ACTIVELY_REDUCE": "Reducing unsecured debt will restore flexibility faster than spreading payments.",
        "AVOID_ADDING": "Avoid adding new debt until buffer improves.",
        "NEEDS_MONITORING": "Monitor EMI ratio and avoid new commitments.",
        "SAFE_TO_HOLD": "Current structure is stable; keep monitoring.",
    }
    return m.get(treatment, "Review your debt mix and prepayment strategy.")


def build_pdf_report(state: dict) -> bytes:
    """Build a simple PDF snapshot of the current state for offline analysis."""
    buffer = BytesIO()
    c = canvas.Canvas(buffer, pagesize=A4)
    width, height = A4
    text = c.beginText(40, height - 40)
    text.setFont("Helvetica", 10)

    def line(s: str = ""):
        nonlocal text, c
        text.textLine(s)
        if text.getY() < 40:
            c.drawText(text)
            c.showPage()
            text = c.beginText(40, height - 40)
            text.setFont("Helvetica", 10)

    line("DebtEase — Analysis Snapshot")
    line("----------------------------------------")
    lm = state["loan_metrics"]
    line("Loan Metrics:")
    line(f"  Total EMI (month 1): ₹{lm['total_emi']:.0f}")
    if lm["debt_free_month"]:
        line(f"  Debt-free in (strategy): {lm['debt_free_month']} months")
    if lm.get("baseline_debt_free_month"):
        line(f"  Debt-free in (baseline): {lm['baseline_debt_free_month']} months")
    line(f"  Flexibility buffer (current): ₹{lm['flexibility_buffer']:.0f}")
    line(f"  Interest saved vs baseline: ₹{lm['interest_saved']:.0f}")
    line()

    cl = state["classification"]
    line("Classification:")
    line(f"  Treatment: {cl['treatment']}")
    line(f"  Confidence: {cl['confidence']}")
    if cl["flags"]:
        line("  Flags:")
        for f in cl["flags"]:
            line(f"    - {f}")
    m = cl["metrics"]
    line("  Metrics:")
    line(f"    EMI ratio: {m['emi_ratio']:.3f}")
    line(f"    Fragmentation index: {m['fragmentation_index']:.3f}")
    line(f"    Rigidity score: {m['rigidity_score']:.3f}")
    line(f"    Shock buffer (months): {m['shock_buffer_months']:.3f}")
    line(f"    Unsecured ratio: {m['unsecured_ratio']:.3f}")
    line()

    stx = state["stress"]
    line("Stress & Shock:")
    line(f"  Stress level (worst preset): {stx['stress_level']}")
    line(f"  Flexibility after worst shock: ₹{stx['flexibility_after_shock']:.0f}")
    line(f"  Survivability band: {stx['survivability_band']}")
    if stx["failed_debt_categories"]:
        line("  Failure order:")
        for cat in stx["failed_debt_categories"]:
            line(f"    - {cat}")
    if stx["key_drivers"]:
        line("  Key drivers:")
        for d in stx["key_drivers"]:
            line(f"    - {d}")
    line()

    pre = state["prepayment"]
    line("Prepayment Strategy:")
    line(f"  Strategy: {pre['strategy']}")
    line(f"  Monthly extra: ₹{pre['monthly_extra']:.0f}")

    c.drawText(text)
    c.showPage()
    c.save()
    buffer.seek(0)
    return buffer.getvalue()

# Priority list: unsecured first, then by balance (snowball) or rate (avalanche)
def build_priority_list(loans_list, schedule, strategy_type: str) -> list:
    month1 = schedule.get(1, [])
    by_id = {r.loan_id: (r.emi, r.closing_balance) for r in month1}
    loan_by_id = {l.loan_id: l for l in loans_list}
    # Sort: unsecured first, then by balance (snowball) or rate (avalanche)
    def key(item):
        loan = loan_by_id[item[0]]
        emi, bal = item[1]
        sec = 0 if not loan.secured else 1
        if strategy_type == StrategyType.SNOWBALL:
            return (sec, bal, -getattr(loan, "annual_rate", 0))
        return (sec, -getattr(loan, "annual_rate", 0), bal)
    ordered = sorted(by_id.items(), key=key)
    return [loan_by_id[lid].instrument_type.replace("_", " ").title() for lid, _ in ordered]

priority_list = build_priority_list(loans, strategy_schedule, strategy)

# =====================
# 3. MAIN SCREEN — Tabs
# =====================
tab_snapshot, tab_debt_health, tab_stress, tab_strategy, tab_what_if = st.tabs([
    "Snapshot",
    "Debt Health",
    "Stress Test",
    "Strategy",
    "What If",
])

# ----- TAB 1: Snapshot -----
with tab_snapshot:
    st.subheader("Snapshot")
    st.caption("Simulated home screen.")

    c1, c2, c3, c4 = st.columns(4)
    c1.metric("Debt Health", debt_health_label(state["stress"]["stress_level"]))
    c2.metric("Monthly EMI", f"₹{state['loan_metrics']['total_emi']:,.0f}")
    c3.metric("Debt freedom", f"{state['loan_metrics']['debt_free_month']} months" if state["loan_metrics"]["debt_free_month"] else "—")
    c4.metric("Flexibility buffer", f"₹{state['loan_metrics']['flexibility_buffer']:,.0f}")

    st.write(snapshot_explanation(state["stress"]["stress_level"], state["classification"]["flags"]))

    st.markdown("---")
    st.subheader("Export")
    if REPORTLAB_AVAILABLE:
        pdf_bytes = build_pdf_report(state)
        st.download_button(
            label="Download PDF report",
            data=pdf_bytes,
            file_name="debtease_analysis.pdf",
            mime="application/pdf",
        )
    else:
        st.info("PDF export requires the 'reportlab' package. Install with: pip install reportlab")

# ----- TAB 2: Debt Health -----
with tab_debt_health:
    st.subheader("Why your debt feels heavy")
    flags = state["classification"]["flags"]
    if flags:
        for f in flags:
            st.write(f"• {flag_to_bullet(f)}")
    else:
        st.write("No major risk flags from the classification engine.")

    with st.expander("Show analysis"):
        m = state["classification"]["metrics"]
        st.write("**EMI ratio** ", m["emi_ratio"])
        st.write("**Fragmentation index** ", m["fragmentation_index"])
        st.write("**Rigidity score** ", m["rigidity_score"])
        st.write("**Shock buffer (months)** ", m["shock_buffer_months"])
        st.write("**Unsecured ratio** ", m["unsecured_ratio"])

# ----- TAB 3: Stress Test -----
with tab_stress:
    st.subheader("Stress Test")
    shock_pct = st.slider("Income shock", 0, 30, 20, help="Simulate income drop %")
    alpha = 1.0 - (shock_pct / 100.0)

    stress_at_shock = simulate_stress(
        income=monthly_income,
        fixed_expenses=fixed_expenses,
        schedule=strategy_schedule,
        loans=loans,
        month=analysis_month,
        alphas=(alpha,),
    )

    st.write(f"**After {shock_pct}% income drop**")
    st.metric("Debt Health", debt_health_label(stress_at_shock.stress_level))
    st.metric("Flexibility after shock", f"₹{stress_at_shock.flexibility_after_shock:,.0f}")

    if stress_at_shock.failed_debt_categories:
        st.write("**First pressure point:**")
        st.write(stress_at_shock.failed_debt_categories[0] if stress_at_shock.failed_debt_categories else "—")
    if stress_at_shock.key_drivers:
        st.caption("Drivers: " + ", ".join(stress_at_shock.key_drivers))

# ----- TAB 4: Strategy -----
with tab_strategy:
    st.subheader("Recommended strategy")
    st.write(f"**{state['classification']['treatment'].replace('_', ' ')}**")
    st.write(strategy_explanation(state["classification"]["treatment"]))

    st.write("**Priority**")
    for i, name in enumerate(priority_list, 1):
        st.write(f"{i}. {name}")

# ----- TAB 5: What If -----
with tab_what_if:
    st.subheader("What If")
    prepayment_slider = st.slider(
        "Prepayment amount (₹/month)",
        0.0,
        200_000.0,
        float(monthly_extra),
        step=1000.0,
    )

    whatif_schedule = simulate_with_strategy(
        loans,
        strategy=strategy,
        monthly_extra=prepayment_slider,
    )
    whatif_free = debt_free_month(whatif_schedule)
    baseline_free = state["loan_metrics"].get("baseline_debt_free_month")
    months_improvement = (baseline_free - whatif_free) if (baseline_free and whatif_free) else None

    whatif_class_input = schedule_to_classification_input(
        whatif_schedule,
        loans,
        analysis_month,
        monthly_net_income=monthly_income,
        fixed_obligations=fixed_expenses,
        extra_short_term_obligations=extra_bnpl_count,
    )
    whatif_class = classify(whatif_class_input)
    whatif_stress = simulate_stress(
        income=monthly_income,
        fixed_expenses=fixed_expenses,
        schedule=whatif_schedule,
        loans=loans,
        month=analysis_month,
    )

    total_emi_whatif = sum(r.emi for r in whatif_schedule.get(1, []))
    emi_ratio_whatif = (total_emi_whatif / monthly_income * 100) if monthly_income else 0

    st.metric("Debt freedom improvement", f"{months_improvement} months earlier" if months_improvement is not None else "—")
    st.metric("New EMI ratio", f"{emi_ratio_whatif:.0f}%")
    st.metric("Stress level", debt_health_label(whatif_stress.stress_level))

# =====================
# DEBUG PANEL
# =====================
st.divider()
debug_mode = st.checkbox("Debug Mode", value=False)
if debug_mode:
    # Serialize state for JSON (drop non-serializable schedule data or summarize)
    debug_state = {
        "loan_metrics": state["loan_metrics"],
        "classification": {
            "treatment": state["classification"]["treatment"],
            "flags": state["classification"]["flags"],
            "confidence": state["classification"]["confidence"],
            "metrics": state["classification"]["metrics"],
        },
        "stress": state["stress"],
        "prepayment": state["prepayment"],
    }
    st.json(debug_state)
