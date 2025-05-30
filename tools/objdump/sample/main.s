	.section	__TEXT,__text,regular,pure_instructions
	.build_version macos, 15, 0
	.globl	_main                           ; -- Begin function main
	.p2align	2
_main:                                  ; @main
	.cfi_startproc
; %bb.0:
	sub	sp, sp, #32
	stp	x29, x30, [sp, #16]             ; 16-byte Folded Spill
	add	x29, sp, #16
	.cfi_def_cfa w29, 16
	.cfi_offset w30, -8
	.cfi_offset w29, -16
	mov	w8, #2                          ; =0x2
	stur	w8, [x29, #-4]
	adrp	x8, _a@PAGE
	ldr	w8, [x8, _a@PAGEOFF]
	ldur	w9, [x29, #-4]
	add	w8, w8, w9
	str	w8, [sp, #8]
	ldr	w8, [sp, #8]
                                        ; kill: def $x8 killed $w8
	mov	x9, sp
	str	x8, [x9]
	adrp	x0, l_.str@PAGE
	add	x0, x0, l_.str@PAGEOFF
	bl	_printf
	mov	w0, #0                          ; =0x0
	ldp	x29, x30, [sp, #16]             ; 16-byte Folded Reload
	add	sp, sp, #32
	ret
	.cfi_endproc
                                        ; -- End function
	.section	__DATA,__data
	.globl	_a                              ; @a
	.p2align	2, 0x0
_a:
	.long	1                               ; 0x1

	.section	__TEXT,__cstring,cstring_literals
l_.str:                                 ; @.str
	.asciz	"\350\256\241\347\256\227\347\273\223\346\236\234\357\274\232%d\n"

.subsections_via_symbols
