# FIDES reliability library for Go

This library is based on FIDES edition 2022. It is not a comprehensive implementation (see the notes below).

Although FIDES and all reliability methods derived from or similar to MIL-HDBK-217 are currently not accurate
according to reports from different sources, NASA included (see the references), some kind of failure estimation
for electronic devices is necessary. FIDES is considered here as a starting point, and as models evolve,
eventually also this library will do.

FIDES adopts a constant failure rate for components but not for subassemblies, which is a contradiction. In future
releases, it is to be expected that more failure models will adopt a Weibull or some other distribution
that more accurately reflects the life expectancy under the given mission profiles.

The library is MIT-licensed (versions before v0.1.0 were BSD-2-Clause) and implements FIDES 2022
only. Since v0.1.0 it includes the FIDES 2022 fixes and the validation made for fitcalc
(github.com/rveen/fitcalc), which uses it as its reliability engine.

## Install

    # go install github.com/rveen/fides/cmd/fides@latest
    # git clone https://github.com/rveen/fides
    # cd fides
    # fides testdata/bom.csv testdata/db.csv testdata/mission.csv

This will do a FIT calculation on the sample BOM provided and print it on screen.

## Input file formats

The fides command (cmd/fides) reads several CSV files, from which the first one is the BOM and the
last one is the mission profile. Any other files augment the data in the BOM, as defined
by 'type' relations (see https://github.com/rveen/golib/csv). The 'name' attribute in BOM files
is the component reference (R1, C4, etc) and should be unique.

BOM items have the following fields:

- 'type': optional, useful to inherit data from other files
- 'class': see next section.
- 'tags': see next section.
- 'value'
- 'package': see "Packages" below.
- 'ndevices': for components that have more than one device per package.
- 'npins': for ICs, optocouplers and connectors (number of contacts).
- 'layers', 'mounts', 'piclass', 'pitechno': for PCBs, the number of layers, the number of
  mounted component terminations, and ΠClass and ΠTechnology (FIDES 2022, pp. 174-176).
- 'tmax': maximum working temperature
- 'vmax': maximum permanent voltage
- 'vpmax': maximum transient voltage (not used at the moment)
- 'pmax': power rating
- 'description': optional field
- 'v': working voltage
- 'i': working current (optional)
- 'p': working power (optional)

The last file to be specified on the command line is the mission profile. 

See [testdata](testdata) for some CSV examples.

## Class and tags

Components are identified by the fields 'class' and 'tags'. Class
takes the values L, C, R, D, Q, U, X or PCB. Tags identify types within a class:

- All: smd (default), tht (for through hole), analog, interface, power (placement factor)
- C / Aluminium electrolytic capacitors: alu, elco; solid (also dry, polymer) for solid electrolyte
- C / Tantalum capacitors: tant, tantalum, tantalium; solid (default: SMD, or axial, bead),
  wet (default: silver case and glass seal; elastomer, glass_sealed, silver_case, tantalum_case)
- C / Ceramic capacitors: cer, x5r, x5s, x6r, x6s, x7r, x7s, x8r, x8s, np0, c0g, y5v (or type1,
  type2); flex for flexible (polymer) terminations; topend for top-end technology (category 3
  above the category 2 C·V limit)
- L / Inductors, transformers: trafo, power, multilayer/ferrite_bead (default: low-current wirewound)
- R / Resistors: thin film (default, by value), thick, melf, network (with ndevices), ww (for
  wirewound; power rating ≥ 1 W selects the power rows), pot/potmeter/potentiometer/variable
- D / Diodes: zener, tvs, rectifier, led (with white, ceramic)
- Q / Transistors: mos/mosfet, jfet, igbt, triac, thyristor (bipolar by default). MOS over 5 W
  and IGBTs use the power transistor model.
- U / ICs: digital, analog, mixed, microprocessor/microcontroller/dsp, dram, sram,
  fpga/cpld/pal, flash/eprom/eeprom
- U / Optocouplers: opto, optocoupler; photodiode (default: phototransistor)
- X / Crystals, resonators; osc/oscillator for oscillators
- J / Connectors for printed circuits: pressfit, tht (default: smd)
- PCB / see the PCB fields above

If the assembly style is not defined (smd or tht), then smd is assumed.

## Packages

IC packages are written as the package family followed by the number of pins (LQFP64,
PBGA256), or the family alone with 'npins'. Families (FIDES 2022, pp. 125-127): PDIP,
CERDIP/CDIP, PQFP, LQFP/TQFP, HQFP/RQFP/POWERQFP, CERPACK, CQFP, PLCC, JLCC/JCLCC, CLCC, SOJ,
SO/SOIC, TSOP, SSOP/QSOP, TSSOP/MSOP/VSSOP/HTSSOP, QFN/DFN/VQFN, QFNFP (QFN with 0.4 mm pitch),
LGA, PBGA (1.27 mm pitch), PBGABT (1.00 mm pitch), POWERBGA/TBGA/SBGA, FCBGA, CBGA, DBGA,
CCGA/CICGA and CPGA.

Discrete semiconductor packages use their usual names (SOT23, SOD123, SMB, DPAK, TO220, …) and
are grouped as in the guide (pp. 135-136). QFN has no default thermal resistance (the guide's
formula needs the package area), so give 'rtha' for QFN parts.

## Notes on this implementation

- Values and formulas follow the FIDES Guide 2022 Edition A (July 2023). Every model and table
  row implemented here is checked against an independent calculation from the guide
  ([testdata/validation](testdata/validation) and guide_cases_test.go).
- Unsupported components:
  - ASICs (the 'complex' tag uses the microprocessor row instead)
  - Film capacitors
  - Fuses
  - Relays
  - Switches
  - GaN and GaAs (RF and microwave) components, and other microwave components
  - Fine-pitch BGA and wafer-level packages
  - COTS
  - Hybrids
  - Subassemblies
  - Batteries
  - Fans
  - Deep sub-micron components

- Process factors are set to default values:
  - 𝚷Ruggedized = 1.7
  - 𝚷PM = 1.7 for active components (ICs, discrete semiconductors, LEDs, optocouplers),
    1.6 for the others
  - 𝚷Process = 4
  - 𝚷LF = 1

- The parts count method is not implemented

## References

- http://fides-reliability.org/
- https://www.sciencedirect.com/science/article/pii/S2772671123002486
- https://nepp.nasa.gov/docs/etw/2019/0619WED/1630%20-%20Bourbouse%20-%20NEPP_2019_FIDES.pdf
- https://calce.umd.edu/event/18917/elseviers-e-prime-presentation-by-dr-diganta-das-assessment-of-the-fides-guide-2022
- https://ieeexplore.ieee.org/document/9319134
- https://www.lamar.edu/engineering/_files/documents/mechanical/dr.-fan-publications/2008/Fan%202008_13%20ECTC_3.pdf


