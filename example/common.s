section .text
@str2long:
	mov edx,@str_2long_data_len ; 替换为常量26
	mov ecx,@str_2long_data ; 替换为数据段第一行字符串地址
	mov ebx, 1
	mov eax, 4
	int 128
	mov ebx, 0
	mov eax, 1
	int 128
	ret
@procBuf:
	mov esi,@buffer
	mov edi,0
	mov ecx,0
	mov eax,0
	mov ebx,10
@cal_buf_len:
	mov cl,[esi+edi]
	cmp ecx,10
	je @cal_buf_len_exit
	inc edi
	imul ebx
	add eax,ecx
	sub eax,48
	jmp @cal_buf_len
@cal_buf_len_exit:
	mov ecx,edi
	mov [@buffer_len],cl
	mov bl,[esi]
	ret
global _start
_start:
	call main
	mov ebx, 0
	mov eax, 1
	int 128
section .data
	@str_2long_data db "字符串长度溢出！",10,13 ; 双字，带换行符
	@str_2long_data_len equ 26 ; 常量，26
	@buffer times 255 db 0 ; 缓冲区 重复 255 个双字，共计 510 字节
	@buffer_len db 0 ; 缓冲区长度 默认 0
	@s_esp dd @s_base
	@s_ebp dd 0
section .bss
	@s_stack times 65536 db 0
@s_base:
