# FIDES reliability library for Go

This library is based on FIDES edition 2022. It is not a comprehensive implementation (see the notes below).

Although FIDES and all reliability methods derived from or similar to MIL-HDBK-217 are currently not accurate
according to reports from different sources, NASA included (see the references), some kind of failure estimation
for electronic devices is necessary. FIDES is considered here as a starting point, and as models evolve,
eventually also this library will do.

FIDES adopts a constant failure rate for components but not for subassemblies, which is a contradiction. In future
releases, it is to be expected that more failure models will adopt a Weibull or some other distribution
that more accurately reflects the life expectancy under the given mission profiles.

## Install

    # git clone https://github.com/rveen/fides
    # cd fides/cmd/fides
    # go build
    # ./fides bom.csv db.csv mission.csv

This will do a FIT calculation on the sample BOM provided and print it on screen.

## Input file formats

Input files are in CSV format, with the first row containing the field names.
The examples given are self-explanatory.

## Class and tags

Components are identified by the fields 'class' and 'tags'. Class
takes the values L, C, R, D, Q, U, X and PCB. Tags identify types within a class:

- All: smd (default), tht, analog, interface, power
- C / Electrolithic capacitors: alu, elco
- C / Tantalium capacitors: tant, tantalium
- L / Inductors, transformers: trafo, power, multilayer/ferrite
- C / Ceramic capacitors: cer, x5r, x5s, x6r, x6s, x7r, x7s, x8r, x8s, np0, c0g, y5v
- R / Resistors: ww, melf, pot/potmeter, thick
- D / Diodes: zener, tvs
- Q / Transistors: gaas, gan, mos/mosfet, jfet, igbt, triac, thyristor
- U / ICs, ASICs: digital, analog, mixed, complex, dram, sram, fpga/cpld/pal, flash/eprom/eeprom
- U / Optocouplers: opto, optocoupler, photodiode, phototrasistor
- X / Resonators
- PCB / fr4 (default)

If the assembly style is not defined (smd or tht), then smd is assumed.

## Notes on this implementation

- ASICs are treated as normal ICs (handled through tags: complex, analog, digital)

- Unsupported components:
  - COTS
  - LEDs
  - Fuses
  - Piezos, crystals, resonators, oscillators
  - Relays
  - Switches
  - Hybrids
  - Microwave components
  - Subassemblies
  - Batteries
  - Fans
  - Deep sub-micron components

- Process factors are set to default values:
  - 𝚷Ruggedized = 1.7
  - 𝚷PM = 1.7
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


