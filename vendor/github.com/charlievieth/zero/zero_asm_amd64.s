#include "textflag.h"

TEXT ·Zero(SB), NOSPLIT, $0-25
	MOVQ a_len+8(FP), BX
	MOVQ a+0(FP), SI
	LEAQ ret+24(FP), AX
	JMP  zero·memzerobody(SB)

// a in SI
// count in BX
// address of result byte in AX
TEXT zero·memzerobody(SB), NOSPLIT, $0-0
	CMPQ BX, $8
	JB   small
	CMPQ BX, $64
	JB   zero_bigloop
	CMPB runtime·support_avx2(SB), $1
	JE   zero_hugeloop_avx2

zero_hugeloop:
	// 64 bytes at a time using xmm registers
	PXOR X1, X1
	PXOR X3, X3
	PXOR X5, X5
	PXOR X7, X7

hugeloop:
	CMPQ     BX, $64
	JB       zero_bigloop
	MOVOU    (SI), X0
	MOVOU    16(SI), X2
	MOVOU    32(SI), X4
	MOVOU    48(SI), X6
	PCMPEQB  X1, X0
	PCMPEQB  X3, X2
	PCMPEQB  X5, X4
	PCMPEQB  X7, X6
	PAND     X2, X0
	PAND     X6, X4
	PAND     X4, X0
	PMOVMSKB X0, DX
	ADDQ     $64, SI
	SUBQ     $64, BX
	CMPL     DX, $0xffff
	JEQ      hugeloop
	MOVB     $0, (AX)
	RET

// 64 bytes at a time using ymm registers
zero_hugeloop_avx2:
	VPXOR Y1, Y1, Y1
	VPXOR Y3, Y3, Y3
	JE    hugeloop_avx2

hugeloop_avx2:
	CMPQ      BX, $64
	JB        bigloop_avx2
	VMOVDQU   (SI), Y0
	VMOVDQU   32(SI), Y2
	VPCMPEQB  Y1, Y0, Y4
	VPCMPEQB  Y2, Y3, Y5
	VPAND     Y4, Y5, Y6
	VPMOVMSKB Y6, DX
	ADDQ      $64, SI
	SUBQ      $64, BX
	CMPL      DX, $0xffffffff
	JEQ       hugeloop_avx2
	VZEROUPPER
	MOVB      $0, (AX)
	RET

bigloop_avx2:
	VZEROUPPER

zero_bigloop:
	// 8 bytes at a time using 64-bit register
	XORQ DX, DX
	JE   bigloop

bigloop:
	CMPQ BX, $8
	JBE  leftover
	MOVQ (SI), CX
	ADDQ $8, SI
	SUBQ $8, BX
	CMPQ CX, DX
	JEQ  bigloop
	MOVB $0, (AX)
	RET

// remaining 0-8 bytes
leftover:
	MOVQ  -8(SI)(BX*1), CX
	XORQ  DX, DX
	CMPQ  CX, DX
	SETEQ (AX)
	RET

small:
	CMPQ BX, $0
	JEQ  equal

	LEAQ 0(BX*8), CX
	NEGQ CX

	CMPB SI, $0xf8
	JA   si_high

	// load at SI wont cross a page boundary.
	MOVQ (SI), SI
	JMP  si_finish

si_high:
	// address ends in 11111xxx. Load up to bytes we want, move to correct position.
	MOVQ -8(SI)(BX*1), SI
	SHRQ CX, SI

si_finish:
	// NEW
	XORQ DI, DI

di_finish:
	SUBQ SI, DI
	SHLQ CX, DI

equal:
	SETEQ (AX)
	RET
