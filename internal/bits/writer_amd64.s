//go:build !purego

#include "textflag.h"

TEXT clear<>(SB), NOSPLIT, $0
    VPXOR Y0, Y0, Y0
    VMOVDQU Y0, (AX)
    VMOVDQU Y0, 32(AX)
    VMOVDQU Y0, 64(AX)
    VMOVDQU Y0, 96(AX)
    VMOVDQU Y0, 128(AX)
    VMOVDQU Y0, 160(AX)
    VMOVDQU Y0, 192(AX)
    VMOVDQU Y0, 224(AX)
    VMOVDQU Y0, 256(AX)
    VMOVDQU Y0, 288(AX)
    VMOVDQU Y0, 320(AX)
    VMOVDQU Y0, 352(AX)
    VMOVDQU Y0, 384(AX)
    VMOVDQU Y0, 416(AX)
    VMOVDQU Y0, 448(AX)
    VMOVDQU Y0, 480(AX)
    VZEROUPPER
    RET

DATA bitmasks8<>+0(SB)/8, $0x0101010101010101
DATA bitmasks8<>+8(SB)/8, $0x0303030303030303
DATA bitmasks8<>+16(SB)/8, $0x0707070707070707
DATA bitmasks8<>+24(SB)/8, $0x0F0F0F0F0F0F0F0F
DATA bitmasks8<>+32(SB)/8, $0x1F1F1F1F1F1F1F1F
DATA bitmasks8<>+40(SB)/8, $0x3F3F3F3F3F3F3F3F
DATA bitmasks8<>+48(SB)/8, $0x7F7F7F7F7F7F7F7F
DATA bitmasks8<>+56(SB)/8, $0xFFFFFFFFFFFFFFFF
GLOBL bitmasks8<>(SB), RODATA|NOPTR, $64

// func writeInt8x8(dst *[512]byte, src [][8]int8, bitWidth uint) int
TEXT ·writeInt8x8(SB), NOSPLIT, $-48
    MOVQ bitWidth+32(FP), DI

    XORQ DX, DX
    MOVQ $512, AX
    IDIVQ DI
    MOVQ AX, CX

    MOVQ dst+0(FP), AX
    MOVQ src_base+8(FP), BX
    MOVQ src_len+16(FP), DX
    CALL clear<>(SB)
    CMPQ CX, DX
    CMOVQLT CX, DX
    MOVQ DX, ret+40(FP)
    SHLQ $3, DX
    ADDQ BX, DX

    LEAQ bitmasks8<>(SB), SI
    MOVQ -8(SI)(DI*8), CX

    CMPQ BX, DX
    JE done
loop:
    MOVQ (BX), SI
    PEXTQ CX, SI, SI
    ORQ SI, (AX)

    ADDQ DI, AX
    ADDQ $8, BX
    CMPQ BX, DX
    JNE loop
done:
    RET
