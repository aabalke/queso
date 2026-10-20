#include "queso.h"
#include <stdint.h>
#include <stdio.h>

int main(void) {

  queso_isa_opts opts = {
      .arch = QUESO_ARMv5TE,
      .order = QUESO_LITTLE_ENDIAN,
      .thumb = false,
      .output_address = true,
      .output_opcode = true,
      .offset = 0xDEADBEEF,
      .undefined_output = "unknown instruction",
      .naming = QUESO_CUSTOM,
      .custom_names =
          {
              "zero",
              "one",
          },
  };

  uint8_t bytes[8] = {
      0x0E, 0xF0, 0xB0, 0xE1, 0x91, 0x00, 0x02, 0xE1,
  };

  return decode(&opts, bytes, sizeof(bytes));
}
