#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

typedef enum {
  QUESO_ARMv1,
  QUESO_ARMv2,
  QUESO_ARMv2a,
  QUESO_ARMv3,
  QUESO_ARMv3M,
  QUESO_ARMv4xM,
  QUESO_ARMv4,
  QUESO_ARMv4TxM,
  QUESO_ARMv4T,
  QUESO_ARMv5xM,
  QUESO_ARMv5,
  QUESO_ARMv5TxM,
  QUESO_ARMv5T,
  QUESO_ARMv5TExP,
  QUESO_ARMv5TE,
} Arch;

typedef enum {
  QUESO_LITTLE_ENDIAN,
  QUESO_BIG_ENDIAN,
  QUESO_NATIVE_ENDIAN,
} Order;

typedef enum {
  QUESO_CONVENTIONAL,
  QUESO_GENERAL,
  QUESO_CUSTOM,
} RegisterNaming;

typedef struct {
  const char *undefined_output;
  const char *custom_names[16];
  RegisterNaming naming;
  Arch arch;
  Order order;
  uint32_t offset;
  bool thumb;
  bool output_address;
  bool output_opcode;
} queso_isa_opts;
