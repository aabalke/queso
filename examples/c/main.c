#include "../../build/c/queso.h"
#include <stdint.h>
#include <stdio.h>

// requires compiling with libqueso.h, queso.h and queso.a
// e. g. (gcc main.c ../../build/c/libqueso.h ../../build/c/queso.a -o main.exe)
// after building c version

int main(void) {

  // all important types and enums are in libqueso.h

  // fill in queso decoding options
  queso_isa_opts opts = {
      .arch = QUESO_ARMv5TE,
      .order = QUESO_LITTLE_ENDIAN,
      //.thumb = false,
      .output_address = true,
      .output_opcode = true,
      //.offset = 0x0,
      .undefined_output = "unknown instruction",
      //.naming = QUESO_CUSTOM,
      //.custom_names =
      //    {
      //        "zero",
      //        "one",
      //    },
  };

  // provide instructions as a byte array
  uint8_t bytes[8] = {
      0x0E, 0xF0, 0xB0, 0xE1, 0x91, 0x00, 0x02, 0xE1,
  };

  // create a result struct for queso to return instruction strings, and errors
  queso_result res = {};

  int ok = queso_decode(&opts, bytes, sizeof(bytes), &res);

  if (ok < 0) {
    printf("error: %s\n", res.err);
    queso_free_result(&res);
    return 0;
  }

  for (int i = 0; i < res.len; i++) {
    printf("%s\n", res.lines[i]);
  };

  queso_free_result(&res);
  return 0;
}
