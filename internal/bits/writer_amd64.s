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

// func writeInt8x8(dst *[512]byte, src [][8]int8, bitWidth uint) int
TEXT ·writeInt8x8(SB), NOSPLIT, $-48
    MOVQ bitWidth+32(FP), DI

    XORQ DX, DX
    MOVQ $512, AX
    IDIVQ DI
    MOVQ AX, SI

    MOVQ dst+0(FP), AX
    MOVQ src_base+8(FP), BX
    MOVQ src_len+16(FP), DX
    CALL clear<>(SB)

    CMPQ SI, DX
    CMOVQLT SI, DX
    MOVQ DX, ret+40(FP)
    SHLQ $3, DX
    ADDQ BX, DX

    MOVQ DI, CX
    MOVQ $1, SI
    SHLQ CX, SI
    DECQ SI
loop:
    CMPQ BX, DX
    JE done

    MOVQ (BX), R8
    MOVQ R8, R9
    MOVQ R8, R10
    MOVQ R8, R11
    MOVQ R8, R12
    MOVQ R8, R13
    MOVQ R8, R14
    MOVQ R8, R15

    SHRQ $8, R9
    SHRQ $16, R10
    SHRQ $24, R11
    SHRQ $32, R12
    SHRQ $40, R13
    SHRQ $48, R14
    SHRQ $56, R15

    ANDQ SI, R8
    ANDQ SI, R9
    ANDQ SI, R10
    ANDQ SI, R11
    ANDQ SI, R12
    ANDQ SI, R13
    ANDQ SI, R14
    ANDQ SI, R15

    MOVQ DI, CX
    SHLQ CX, R9
    ADDQ DI, CX
    SHLQ CX, R10
    ADDQ DI, CX
    SHLQ CX, R11
    ADDQ DI, CX
    SHLQ CX, R12
    ADDQ DI, CX
    SHLQ CX, R13
    ADDQ DI, CX
    SHLQ CX, R14
    ADDQ DI, CX
    SHLQ CX, R15

    ORQ R15, R14
    ORQ R13, R12
    ORQ R11, R10
    ORQ R9, R8
    ORQ R14, R12
    ORQ R10, R8
    ORQ R12, R8
    ORQ R8, (AX)

    ADDQ DI, AX
    ADDQ $8, BX
    JMP loop
done:
    RET
