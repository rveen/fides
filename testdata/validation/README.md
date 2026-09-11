# FIDES 2022 validation

The library's FIT values are checked against an independent calculation written
from the FIDES Guide 2022 Edition A (July 2023), not from this Go package. Page
numbers in the scripts refer to that guide.

## Method

One case per model and table row, for the mission profile `testdata/mission.csv`
(the `cmd/fides` sample) unless noted:

- `cases.csv` has 99 components, each naming the guide model and table row
  that applies (`model`), its placement factor and its mission profile.
- `fides_cases.py` computes them from the guide, reusing the common factors
  of `fides_ref.py`; `reference-cases.csv` holds its results.
- `guide_cases_test.go` (`TestGuideCases`) builds the same components from
  class, tags, package and ratings and requires the library to match within
  10⁻⁹. The library therefore has to pick the model the case names.
- `mission-unsealed.csv` is the same mission with a non-hermetic enclosure,
  so the chemical terms of PCBs and connectors are tested too.

`fides_ref.py` also computes the reference values of fitcalc's probe circuit
(github.com/rveen/fitcalc, `testdata/validation/reference.csv` there), which
check the whole fitcalc pipeline, SPICE stresses included.

Regenerate and compare:

    python3 testdata/validation/fides_cases.py > testdata/validation/reference-cases.csv
    go test .

## Coverage

| Family | Guide pages | Rows covered |
|---|---|---|
| Resistors | 146-148 | thin film (3 value ranges × SMD/through-hole), thick film SMD, high-power thick film (rating ≥ 1 W), MELF, wirewound precision and power, potentiometer, network (√N) |
| Ceramic capacitors | 151-153 | type I and type II, categories 1-3, flexible terminations (X7R and other dielectrics) |
| Aluminium capacitors | 154-155 | liquid electrolyte, solid electrolyte |
| Tantalum capacitors | 156-157 | solid SMD, axial, bead; wet with silver case and glass seal, elastomer seal, tantalum case |
| Magnetic components | 161-162 | wirewound low current and power, multilayer, low-power and power transformers |
| Discrete semiconductors | 133-137 | signal, rectifier, zener, TVS (low and high power), thyristor, bipolar, MOS, JFET, multiple devices (√N); package groups THT signal, SMD signal, SMD medium, THT power, SMD C-lead, SMD power, screw power, glass, metal can, flat lead |
| Power transistors | 138-140 | silicon MOS > 5 W and IGBT (ΠPW = 10, 60 °C reference, Tj capped at 175 °C) |
| LEDs | 141-143 | colour and white (√N), every package row of p. 142 for IF < 150 mA and ≥ 150 mA |
| Optocouplers | 144-145 | phototransistor and photodiode output, multiple channels (√N) |
| Integrated circuits | 123-129 | every chip type of p. 128; PDIP, CERDIP, PQFP, LQFP, power QFP, CERPACK, CQFP, PLCC, JLCC, CLCC, SOJ, SO, TSOP, SSOP, TSSOP, QFN, QFN 0.4 mm, LGA, PBGA, PBGA BT, power BGA, FCBGA, CBGA, DBGA, CCGA, CPGA |
| Crystals and oscillators | 163-164 | resonator and oscillator, SMD and through-hole, ΠratingTH and ΠratingEL |
| Printed circuit boards | 174-176 | sealed and non-hermetic |
| Connectors for printed circuits | 177-180 | press-fit, through-hole, SMD; sealed and non-hermetic |

## Deviations found and fixed

### Found with fitcalc's probe circuit

Before the fixes, the library was 69 % below the guide for resistors and
ceramic capacitors, and up to 6× above it for diodes. It was changed to follow
the guide:

| Deviation | Guide | Effect before the fix |
|---|---|---|
| Cycle duration factor of ΠTCy (solder joints) was (min(θcy,2)/2)^1.3 | ^(1/3), pp. 43-44 | TCy of every family too low (up to ×44 for short cycles) |
| Discrete semiconductor package base rates were not the 2022 values | pp. 135-136 | diodes ×7.9, small transistors +40 % |
| ΠPM default 1.7 for all components | 1.6 for passives, p. 36 | passives +6 % |
| Resistor temperature Tamb + P·Rth of the package | Tamb + A·P/Prated, p. 148 | small here (γTH-EL = 0.01) |
| High-power resistor variants chosen by working power | by rating (≥ 1 W), p. 147 | wrong model above 1 W |
| Analog and mixed IC λ0TH 0.123 | 0.086, p. 128 | analog ICs +43 % on the thermal term |
| QFN 64-80 pins: mechanical a = 9.17, up to 72 pins | 9.64, up to 80 pins, p. 126 | QFN over 56 pins |
| Ceramic type II category 3 γMech 0.02 | 0.05, p. 152 | large-CV ceramic capacitors |
| Environmental pollution "moderate" = 1 | 1.5, p. 180 | connectors in non-hermetic products |
| Signal diode ΠEl floor 0.3^2.4 = 0.0556 | 0.056, p. 137 | < 1 % |
| 1/k_B = 11604.518 | 11604 in the formulas, p. 41 | ≤ 0.03 % |

Found earlier and already fixed: integer constant divisions in the Arrhenius
and ΠTCy formulas, and the transistor base rate.

### Other families

| Deviation | Guide |
|---|---|
| Optocouplers always gave 0 FIT, and ICs tagged `opto` used the IC model | pp. 144-145 |
| LED model: not the guide's package table, chip rates or activation energy | pp. 141-143 |
| Wirewound precision resistor λ0 0.03 | 0.3, p. 147 |
| Resistor networks, potentiometers and high-power rows incomplete; Csensitivity not taken from the model's row | pp. 146-148 |
| Ceramic flexible terminations, type I category 3, aluminium solid electrolyte rows | pp. 151-155 |
| Tantalum defaults and wet variants | pp. 156-157 |
| Csensitivity of magnetic components not taken from their row | p. 162 |
| Crystals: ΠratingTH applied only above Tmax − 40 °C, no ΠratingEL for oscillators, SMD/through-hole choice | pp. 163-164 |
| PCB model unreachable (stopped by the Tmax check), wrong mechanical and humidity coefficients, no layers/mounts/class/technology inputs | pp. 174-176 |
| Power MOS > 5 W and IGBTs: no ΠPW = 10, wrong reference temperature and λ0TH | pp. 138-140 |
| Connector placement factor taken from the tags | fixed at 1, p. 177 |
| IC packages missing (power QFP, CERPACK, CQFP, PLCC, JLCC, CLCC, SOJ, LGA, BGA variants, CCGA, CPGA); SOT matched as TSOP | pp. 125-127 |
| Default junction-to-ambient thermal resistances | p. 119 |

## Not supported

The library returns an error, or the component has no model:

- fuses, relays, switches, film capacitors, ASICs (the `complex` tag uses the
  microprocessor row instead), GaN and GaAs (RF/microwave) components, and the
  other families the readme lists;
- fine-pitch BGA and wafer-level packages;
- a default thermal resistance for QFN (the guide's formula needs the package
  area), so QFN parts need `rth` or `rtha`.

The mission profile gives Πapplication directly (the guide's questionnaire,
p. 113, is not implemented). ΠRuggedizing, ΠPM and ΠProcess are the guide's
defaults.
