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

Input files are in CSV format. The first row contains the field names. If the first
column is 'code' it is considered as a database file, and all fields of each code are applied
to components that in BOMs have that code. BOM files should be loaded (specified in the 
command line) before DB files.

BOM files need only have 2 fields: 'name' and 'code', provided that the other fields needed
are in the database file. The 'name' attribute in BOM files is the component reference 
(R1, C4, etc) and should be unique.

Database files have the following fields:

- 'code'
- 'class': see next section.
- 'tags': see next section.
- 'value'
- 'package'
- 'ndevices': for components that have with more than one device per package.
- 'npins': for ICs.
- 'tmax': maximum working temperature
- 'vmax': maximum permanent voltage
- 'vpmax': maximum transient voltage (not used at the moment)
- 'pmax': power rating
- 'description': optional field.

The last file to be specified on the command line is the mission profile. 

See [here](cmd/fides) for some CSV examples.

## Class and tags

Components are identified by the fields 'class' and 'tags'. Class
takes the values L, C, R, D, Q, U, X or PCB. Tags identify types within a class:

- All: smd (default), tht (for through hole), analog, interface, power
- C / Electrolithic capacitors: alu, elco
- C / Tantalium capacitors: tant, tantalium
- L / Inductors, transformers: trafo, power, multilayer/ferrite_bead
- C / Ceramic capacitors: cer, x5r, x5s, x6r, x6s, x7r, x7s, x8r, x8s, np0, c0g, y5v
- R / Resistors: ww (for wirewound), melf, pot/potmeter, thick
- D / Diodes: zener, tvs
- Q / Transistors: gaas, gan, mos/mosfet, jfet, igbt, triac, thyristor
- U / ICs, ASICs: digital, analog, mixed, complex, dram, sram, fpga/cpld/pal, flash/eprom/eeprom
- U / Optocouplers: opto, optocoupler, photodiode, phototrasistor
- X / Crystals, resonators
- J / pressfit
- PCB / 

If the assembly style is not defined (smd or tht), then smd is assumed.

## Notes on this implementation

- ASICs are treated as normal ICs (handled through tags: complex, analog, digital)
- Current rating in crystals is not implemented

- Unsupported components:
  - COTS
  - LEDs
  - Fuses
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


