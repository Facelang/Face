package link

import (
	"debug/elf"
	"encoding/binary"
	"fmt"
)

const (
	StartSymbol = "_start"   // 程序入口符号
	BaseAddr    = 0x08040000 // 默认加载地址
	MemAlign    = 4096       // 默认内存对齐大小 4KB
	DiscAlign   = 4          // 默认磁盘对齐大小 4B
	// 重定位类型
	R_386_NONE     = 0  // 无重定位
	R_386_32       = 1  // 绝对地址修正
	R_386_PC32     = 2  // 相对地址修正
	R_386_GOT32    = 3  // GOT 入口地址修正
	R_386_PLT32    = 4  // PLT 入口地址修正
	R_386_COPY     = 5  // 复制符号到目标位置
	R_386_GLOB_DAT = 6  // 设置 GOT 入口为符号地址
	R_386_JMP_SLOT = 7  // 设置 PLT 入口为符号地址
	R_386_RELATIVE = 8  // 基址重定位
	R_386_GOTOFF   = 9  // GOT 相对地址修正
	R_386_GOTPC    = 10 // PC 相对 GOT 地址修正
)

// Block 表示一个数据块
type Block struct {
	Data   []byte
	Offset uint32
	Size   uint32
}

// SegList 表示段的列表, 还有两个方法： allocAddr, relocAddr
type SegList struct {
	BaseAddr  uint32     // 分配基地址
	Offset    uint32     // 合并后的文件偏移
	Size      uint32     // 合并后大小
	Begin     uint32     // 对齐前开始位置偏移
	OwnerList []*ElfFile // 拥有该段的文件序列
	Blocks    []*Block   // 记录合并后的数据块序列
}

// AllocAddr 分配地址空间
func (s *SegList) AllocAddr(name string, base *uint32, off *uint32) {
	s.Begin = *off //记录对齐前偏移

	// 虚拟地址对齐，让所有的段按照4KB字节对齐
	if name != ".bss" {
		*base += (MemAlign - *base%MemAlign) % MemAlign
	}

	// 偏移地址对齐，让一般段按照4字节对齐，文本段按照16字节对齐
	align := uint32(DiscAlign)
	if name == ".text" {
		align = 16
	}
	*off += (align - *off%align) % align

	// 使虚址和偏移按照4KB模同余
	*base = *base - *base%MemAlign + *off%MemAlign

	// 累加地址和偏移
	s.BaseAddr = *base
	s.Offset = *off
	s.Size = 0
	for _, file := range s.OwnerList {
		s.Size += (DiscAlign - s.Size%DiscAlign) % DiscAlign // 对齐每个小段，按照4字节
		seg := file.ShdrTab[name]
		//读取需要合并段的数据
		if name != ".bss" {
			data, _ := file.GetData(seg)
			block := &Block{
				Data:   data,
				Offset: s.Size,
				Size:   uint32(len(data)),
			}
			s.Blocks = append(s.Blocks, block)
		}
		//修改每个文件中对应段的addr
		seg.Addr = *base + s.Size //修改每个文件的段虚拟，为了方便计算符号或者重定位的虚址，不需要保存合并后文件偏移
		s.Size += seg.Size        //累加段大小
	}
	*base += s.Size
	if name != ".bss" {
		*off += s.Size
	}
}

// RelocAddr 根据提供的重定位信息重定位地址
func (s *SegList) RelocAddr(relAddr uint32, relocType uint8, symAddr uint32) {
	relOffset := relAddr - s.BaseAddr //同类合并段的数据偏移

	// 查找修正地址所在位置的数据块
	var targetBlock *Block
	for _, block := range s.Blocks {
		if block.Offset <= relOffset && block.Offset+block.Size > relOffset {
			targetBlock = block
			break
		}
	}
	if targetBlock == nil {
		return
	}

	//处理字节为b->data[relOffset-b->offset]
	// 获取需要修改的地址位置
	offset := relOffset - targetBlock.Offset
	if offset+4 > uint32(len(targetBlock.Data)) {
		return
	}

	// 获取当前地址值
	currentAddr := binary.LittleEndian.Uint32(targetBlock.Data[offset:])

	// 根据重定位类型进行修正
	switch relocType {
	case R_386_32: // 绝对地址修正
		binary.LittleEndian.PutUint32(targetBlock.Data[offset:], symAddr)
	case R_386_PC32: // 相对地址修正
		newAddr := symAddr - relAddr + currentAddr
		binary.LittleEndian.PutUint32(targetBlock.Data[offset:], newAddr)
	}
}

// SymLink 表示符号链接， 符号引用对象，每一次外部符号引用都会产生
type SymLink struct {
	Name string
	Recv *ElfFile // 引用符号的文件
	Prov *ElfFile // 提供符号的文件， 符号未定义时必然是NULL
}

// Linker 是链接器的主要结构
type Linker struct {
	SegNames   []string            // 链接关心的段
	Exe        *ElfFile            // 链接后的输出文件
	StartOwner *ElfFile            // 拥有全局符号START/_start的文件
	Elfs       []*ElfFile          // 所有目标文件对象
	SegLists   map[string]*SegList // 所有合并段表序列
	SymLinks   []*SymLink          // 所有符号引用信息，符号解析前存储未定义的符号prov字段为NULL
	SymDef     []*SymLink          // 所有符号定义信息recv字段NULL时标示该符号没有被任何文件引用，否则指向本身（同prov）
	ByteOrder  binary.ByteOrder    // 大端序？小端序？
}

// NewLinker 创建新的链接器
func NewLinker() *Linker {
	l := &Linker{
		SegNames: []string{".text", ".data", ".bss"},
		SegLists: make(map[string]*SegList),
		SymLinks: make([]*SymLink, 0),
		SymDef:   make([]*SymLink, 0),
	}
	// 初始化段列表
	for _, name := range l.SegNames {
		l.SegLists[name] = &SegList{
			OwnerList: make([]*ElfFile, 0),
			Blocks:    make([]*Block, 0),
		}
	}
	return l
}

// AddElf 添加一个目标文件
func (l *Linker) AddElf(name string) error {
	file, err := ReadElf(name)
	if err != nil {
		return fmt.Errorf("failed to open ELF file %s: %v", name, err)
	}
	l.Elfs = append(l.Elfs, file)
	return nil
}

// CollectInfo 收集段信息和符号关联信息
func (l *Linker) CollectInfo() error {
	l.SymLinks = []*SymLink{}
	l.SymDef = []*SymLink{}
	for _, elf := range l.Elfs {
		for _, seg := range l.SegNames {
			if _, ok := elf.ShdrTab[seg]; ok {
				l.SegLists[seg].OwnerList = append(l.SegLists[seg].OwnerList, elf)
			}
		}

		for _, name := range elf.SymNames {
			if name == "" {
				continue
			}
			sym := elf.SymTab[name]
			link := SymLink{Name: name}
			// todo 还可以判断全局符号，这里直接省略
			if sym.Shndx == 0 { // 未定义？什么意思
				link.Recv = elf
				l.SymLinks = append(l.SymLinks, &link)
			} else {
				link.Prov = elf
				l.SymDef = append(l.SymDef, &link)
			}
		}
	}
	return nil
}

// SymValid 验证符号关联
func (l *Linker) SymValid() bool {
	flag := true
	l.StartOwner = nil
	// 检查定义符号，找入口和重定义
	for i, def := range l.SymDef {
		//if(ELF32_ST_BIND(symDef[i]->prov->symTab[symDef[i]->name]->st_info)!=STB_GLOBAL)//只考虑全局符号
		//continue;
		if def.Prov == nil {
			continue
		}
		if def.Prov.SymTab[def.Name].Info>>4 != 1 {
			continue
		}
		if def.Name == StartSymbol { // 记录程序入口文件
			l.StartOwner = def.Prov
		}
		for j := i + 1; j < len(l.SymDef); j++ { //遍历后边定义的符号
			jDef := l.SymDef[j]
			if jDef.Prov == nil {
				continue
			}
			if jDef.Prov.SymTab[jDef.Name].Info>>4 != 1 {
				continue
			}
			if def.Name == jDef.Name { //同名符号
				fmt.Printf("符号名%s在文件%s和文件%s中发生链接冲突。\n", def.Name, def.Prov.File, jDef.Prov.File)
				flag = false
			}
		}
	}
	if l.StartOwner == nil {
		fmt.Printf("链接器找不到程序入口%s。\n", StartSymbol)
		flag = false
	}

	// 检查未定义符号
	for _, link := range l.SymLinks {
		for _, def := range l.SymDef {
			if link.Name == def.Name {
				if def.Prov == nil {
					continue
				}
				if def.Prov.SymTab[def.Name].Info>>4 != 1 {
					continue
				}
				if link.Name == def.Name { //同名符号
					link.Prov = def.Prov
					def.Recv = def.Prov
				}
				break
			}
		}
		if link.Prov == nil { // 未定义
			info := link.Recv.SymTab[link.Name].Info
			iType := "符号"
			if info&0xf == 1 {
				iType = "变量"
			} else if info&0xf == 2 {
				iType = "函数"
			}
			fmt.Printf("文件%s的%s名%s未定义。\n", link.Recv.File, iType, link.Name)
			if flag {
				flag = false
			}
		}
	}
	return flag
}

// AllocAddr 方法
func (l *Linker) AllocAddr() error {
	currentAddr := uint32(BaseAddr)
	currentOffset := uint32(52 + 32*len(l.SegNames)) // 默认文件偏移，PHT保留.bss段

	// 按段类型分配地址
	for _, segName := range l.SegNames {
		if segList, exists := l.SegLists[segName]; exists {
			segList.AllocAddr(segName, &currentAddr, &currentOffset) //自动分配
		}
	}
	return nil
}

// SymParser 解析符号
func (l *Linker) SymParser() error {
	// 扫描所有定义符号，计算虚拟地址
	for _, def := range l.SymDef {
		sym := def.Prov.SymTab[def.Name]                   // 定义的符号信息
		seg := def.Prov.ShdrNames[sym.Shndx]               // 段名
		sym.Value = sym.Value + def.Prov.ShdrTab[seg].Addr // 偏移 + 段基地址
	}
	// 扫描所有符号引用，绑定虚拟地址
	for _, link := range l.SymLinks {
		provsym := link.Prov.SymTab[link.Name] //被引用的符号信息
		recvsym := link.Recv.SymTab[link.Name] //被引用的符号信息
		recvsym.Value = provsym.Value
	}
	return nil
}

// Relocate 重定位
func (l *Linker) Relocate() {
	// 重定位项符号必然在符号表中，且地址已经解析完毕
	for _, elf := range l.Elfs {
		for _, rel := range elf.RelTab {
			sym := elf.SymTab[rel.RelName]                            //重定位符号信息
			symAddr := sym.Value                                      //解析后的符号段偏移为虚拟地址
			relAddr := elf.ShdrTab[rel.SegName].Addr + rel.Rel.Offset //重定位地址
			l.SegLists[rel.SegName].RelocAddr(relAddr, uint8(rel.Rel.Info&0xff), symAddr)
		}
	}
}

// AssemExe 生成 ELF 可执行文件
func (l *Linker) AssemExe() error {
	// 初始化文件头
	l.Exe.Ehdr.Magic = [16]byte{0x7f, 'E', 'L', 'F', 0x1, 0x1}
	//l.Exe.Ehdr.Class = 1      // 32位
	//l.Exe.Ehdr.Data = 1       // 小端
	l.Exe.Ehdr.Version = 1    // 当前版本
	l.Exe.Ehdr.Type = 2
	l.Exe.Ehdr.Machine = 3
	l.Exe.Ehdr.Flags = 0
	l.Exe.Ehdr.Ehsize = 52

	// 数据位置指针
	curOff := uint32(52 + 32*len(l.SegNames)) // 文件头52B+程序头表项32*个数

	// 添加空节头表项
	l.Exe.AddShdr("", 0, 0, 0, 0, 0, 0, 0, 0, 0) //空段表项

	// 计算节名字符串表大小
	shstrtabSize := 26 // ".shstrtab".length()+".symtab".length()+".strtab".length()+3
	for _, name := range l.segNames {
		shstrtabSize += len(name) + 1 // 考虑结束符'\0'
	}

	// 生成程序头表和节头表
	for _, name := range l.segNames {
		seg := l.segLists[name]
		flags := elf.PF_W | elf.PF_R // 读写
		filesz := seg.size

		if name == ".text" {
			flags = elf.PF_X | elf.PF_R // .text段可读可执行
		}
		if name == ".bss" {
			filesz = 0 // .bss段不占磁盘空间
		}

		// 添加程序头表项
		l.exe.AddPhdr(elf.PT_LOAD, seg.offset, seg.baseAddr, filesz, seg.size, flags, elf.MEM_ALIGN)

		// 生成节头表项
		shType := elf.SHT_PROGBITS
		shFlags := elf.SHF_ALLOC | elf.SHF_WRITE
		shAlign := uint32(4) // 4B

		if name == ".bss" {
			shType = elf.SHT_NOBITS
		}
		if name == ".text" {
			shFlags = elf.SHF_ALLOC | elf.SHF_EXECINSTR
			shAlign = 16 // 16B
		}

		l.exe.AddShdr(name, shType, shFlags, seg.baseAddr, seg.offset, seg.size, elf.SHN_UNDEF, 0, shAlign, 0)
		curOff = seg.offset
	}

	// 设置程序头表信息
	l.Exe.Ehdr.Phoff = 52
	l.Exe.Ehdr.Phentsize = 32
	l.Exe.Ehdr.Phnum = uint16(len(l.segNames))

	// 填充节名字符串表
	l.exe.Shstrtab = make([]byte, shstrtabSize)
	index := 0

	// 添加基本节名
	l.shstrIndex[".shstrtab"] = index
	copy(l.exe.Shstrtab[index:], ".shstrtab")
	index += 10

	l.shstrIndex[".symtab"] = index
	copy(l.exe.Shstrtab[index:], ".symtab")
	index += 8

	l.shstrIndex[".strtab"] = index
	copy(l.exe.Shstrtab[index:], ".strtab")
	index += 8

	l.shstrIndex[""] = index - 1

	// 添加其他节名
	for _, name := range l.segNames {
		l.shstrIndex[name] = index
		copy(l.exe.Shstrtab[index:], name)
		index += len(name) + 1
	}

	// 添加.shstrtab节
	l.exe.AddShdr(".shstrtab", elf.SHT_STRTAB, 0, 0, curOff, uint32(shstrtabSize), elf.SHN_UNDEF, 0, 1, 0)
	l.Exe.Ehdr.Shstrndx = uint16(l.exe.GetSegIndex(".shstrtab"))
	curOff += uint32(shstrtabSize)

	// 设置节头表信息
	l.Exe.Ehdr.Shoff = curOff
	l.Exe.Ehdr.Shentsize = 40
	l.Exe.Ehdr.Shnum = uint16(4 + len(l.segNames))

	// 生成符号表
	curOff += 40 * uint32(4+len(l.segNames))
	l.exe.AddShdr(".symtab", elf.SHT_SYMTAB, 0, 0, curOff, uint32((1+len(l.symDef))*16), 0, 0, 1, 16)
	l.exe.GetShdr(".symtab").Link = uint32(l.exe.GetSegIndex(".symtab") + 1)

	// 添加空符号表项
	l.exe.AddSym("", nil)

	// 计算字符串表大小
	strtabSize := 0
	for _, sym := range l.symDef {
		strtabSize += len(sym.name) + 1
	}

	// 添加符号表项
	for _, sym := range l.symDef {
		elfSym := sym.prov.SymTab[sym.name]
		elfSym.Shndx = uint16(l.exe.GetSegIndex(sym.prov.ShdrNames[elfSym.Shndx]))
		l.exe.AddSym(sym.name, elfSym)
	}

	// 设置程序入口点
	l.Exe.Ehdr.Entry = l.exe.SymTab["_start"].Value

	// 添加.strtab节
	curOff += uint32((1 + len(l.symDef)) * 16)
	l.exe.AddShdr(".strtab", elf.SHT_STRTAB, 0, 0, curOff, uint32(strtabSize), elf.SHN_UNDEF, 0, 1, 0)

	// 填充字符串表
	l.exe.Strtab = make([]byte, strtabSize)
	index = 0
	l.strIndex[""] = strtabSize - 1

	for _, sym := range l.symDef {
		l.strIndex[sym.name] = index
		copy(l.exe.Strtab[index:], sym.name)
		index += len(sym.name) + 1
	}

	// 更新符号表名称
	for name, sym := range l.exe.SymTab {
		sym.Name = uint32(l.strIndex[name])
	}

	// 更新节头表名称
	for name, shdr := range l.exe.ShdrTab {
		shdr.Name = uint32(l.shstrIndex[name])
	}

	return nil
}

void Linker::assemExe()
{
//printf("----------------生成elf文件------------------\n");
//初始化文件头
int*p_id=(int*)exe.ehdr.e_ident;
*p_id=0x464c457f;
p_id++;
*p_id=0x010101;
p_id++;
*p_id=0;
p_id++;
*p_id=0;

//for(int i=0;i<EI_NIDENT;++i)printf("%02x ",exe.ehdr.e_ident[i]);printf("\n");
exe.ehdr.e_type=ET_EXEC;
exe.ehdr.e_machine=EM_386;
exe.ehdr.e_version=EV_CURRENT;
exe.ehdr.e_flags=0;
exe.ehdr.e_ehsize=52;
//数据位置指针
unsigned int curOff=52+32*segNames.size();//文件头52B+程序头表项32*个数
//printf("Elf header & Pht(32N):\tbase=00000000\tsize=%08x\n",curOff);
exe.addShdr("",0,0,0,0,0,0,0,0,0);//空段表项
int shstrtabSize=26;//".shstrtab".length()+".symtab".length()+".strtab".length()+3;//段表字符串表大小
for(int i=0;i<segNames.size();++i)
{
string name=segNames[i];
shstrtabSize+=name.length()+1;//考虑结束符'\0'
//生成程序头表
Elf32_Word flags=PF_W|PF_R;//读写
Elf32_Word filesz=segLists[name]->size;//占用磁盘大小
if(name==".text")flags=PF_X|PF_R;//.text段可读可执行
if(name==".bss")filesz=0;//.bss段不占磁盘空间
exe.addPhdr(PT_LOAD,segLists[name]->offset,segLists[name]->baseAddr,
filesz,segLists[name]->size,flags,MEM_ALIGN);//添加程序头表项
//计算有效数据段的大小和偏移,最后一个决定
//printf("%s:\tbase=%08x\tsize=%08x\n",name.c_str(),curOff,segLists[name]->size);
curOff=segLists[name]->offset;//修正当前偏移，循环结束后保留的是.bss的基址

//生成段表项
Elf32_Word sh_type=SHT_PROGBITS;
Elf32_Word sh_flags=SHF_ALLOC|SHF_WRITE;
Elf32_Word sh_align=4;//4B
if(name==".bss")sh_type=SHT_NOBITS;
if(name==".text")
{
sh_flags=SHF_ALLOC|SHF_EXECINSTR;
sh_align=16;//16B
}
exe.addShdr(name,sh_type,sh_flags,segLists[name]->baseAddr,segLists[name]->offset,
segLists[name]->size,SHN_UNDEF,0,sh_align,0);//添加一个段表项，暂时按照4字节对齐
}
exe.ehdr.e_phoff=52;
exe.ehdr.e_phentsize=32;
exe.ehdr.e_phnum=segNames.size();
//填充shstrtab数据
char*str=exe.shstrtab=new char[shstrtabSize];
exe.shstrtabSize=shstrtabSize;
int index=0;
//段表串名与索引映射
hash_map<string,int,string_hash> shstrIndex;
shstrIndex[".shstrtab"]=index;strcpy(str+index,".shstrtab");index+=10;
shstrIndex[".symtab"]=index;strcpy(str+index,".symtab");index+=8;
shstrIndex[".strtab"]=index;strcpy(str+index,".strtab");index+=8;
shstrIndex[""]=index-1;
for(int i=0;i<segNames.size();++i)
{
shstrIndex[segNames[i]]=index;
strcpy(str+index,segNames[i].c_str());
index+=segNames[i].length()+1;
}
//for(int i=0;i<shstrtabSize;++i)printf("%c",str[i]);printf("\n");
//for(int i=0;i<shstrtabSize;++i)printf("%d|",str[i]);printf("\n");
//生成.shstrtab
//printf(".shstrtab:\tbase=%08x\tsize=%08x\n",curOff,shstrtabSize);
exe.addShdr(".shstrtab",SHT_STRTAB,0,0,curOff,shstrtabSize,SHN_UNDEF,0,1,0);//.shstrtab
exe.ehdr.e_shstrndx=exe.getSegIndex(".shstrtab");//1+segNames.size();//空表项+所有段数
curOff+=shstrtabSize;//段表偏移
exe.ehdr.e_shoff=curOff;
exe.ehdr.e_shentsize=40;
exe.ehdr.e_shnum=4+segNames.size();//段表数
//printf("Sht(40N):\tbase=%08x\tsize=%08x\n",curOff,40*exe.ehdr.e_shnum);
//生成符号表项
curOff+=40*(4+segNames.size());//符号表偏移
//printf(".symtab:\tbase=%08x\tsize=%08x\n",curOff,(1+symDef.size())*16);
//符号表位置=（空表项+所有段数+段表字符串表项+符号表项+字符串表项）*40
//.symtab,sh_link 代表.strtab索引，默认在.symtab之后,sh_info不能确定
exe.addShdr(".symtab",SHT_SYMTAB,0,0,curOff,(1+symDef.size())*16,0,0,1,16);
exe.shdrTab[".symtab"]->sh_link=exe.getSegIndex(".symtab")+1;//。strtab默认在.symtab之后
int strtabSize=0;//字符串表大小
exe.addSym("",NULL);//空符号表项
for(int i=0;i<symDef.size();++i)//遍历所有符号
{
string name=symDef[i]->name;
strtabSize+=name.length()+1;
Elf32_Sym*sym=symDef[i]->prov->symTab[name];
sym->st_shndx=exe.getSegIndex(symDef[i]->prov->shdrNames[sym->st_shndx]);//重定位后可以修改了
exe.addSym(name,sym);
}
//记录程序入口
exe.ehdr.e_entry=exe.symTab[START]->st_value;//程序入口地址
curOff+=(1+symDef.size())*16;//.strtab偏移
exe.addShdr(".strtab",SHT_STRTAB,0,0,curOff,strtabSize,SHN_UNDEF,0,1,0);//.strtab
//printf(".strtab:\tbase=%08x\tsize=%08x\n",curOff,strtabSize);
//填充strtab数据
str=exe.strtab=new char[strtabSize];
exe.strtabSize=strtabSize;
index=0;
//串表与索引映射
hash_map<string,int,string_hash> strIndex;
strIndex[""]=strtabSize-1;
for(int i=0;i<symDef.size();++i)
{
strIndex[symDef[i]->name]=index;
strcpy(str+index,symDef[i]->name.c_str());
index+=symDef[i]->name.length()+1;
}
//for(int i=0;i<strtabSize;++i)printf("%c",str[i]);printf("\n");
//for(int i=0;i<strtabSize;++i)printf("%d|",str[i]);printf("\n");
//更新符号表name
for(hash_map<string,Elf32_Sym*,string_hash>::iterator i=exe.symTab.begin();i!=exe.symTab.end();++i)
{
i->second->st_name=strIndex[i->first];
//printf("%s\t%08x\t%s\n",i->first.c_str(),strIndex[i->first],str+strIndex[i->first]);
}
//更新段表name
for(hash_map<string, Elf32_Shdr*,string_hash>::iterator i=exe.shdrTab.begin();i!=exe.shdrTab.end();++i)
{
i->second->sh_name=shstrIndex[i->first];
}
}

void Linker::exportElf(const char*dir)
{
exe.writeElf(dir,1);//输出链接后的elf前半段
//输出重要段数据
FILE*fp=fopen(dir,"a+");
char pad[1]={0};
for(int i=0;i<segNames.size();++i)
{
SegList*sl=segLists[segNames[i]];
int padnum=sl->offset-sl->begin;
while(padnum--)
fwrite(pad,1,1,fp);//填充
//输出数据
if(segNames[i]!=".bss")
{
Block*old=NULL;
char instPad[1]={(char)0x90};
for(int j=0;j<sl->blocks.size();++j)
{
Block*b=sl->blocks[j];
if(old!=NULL)//填充小段内的空隙
{
padnum=b->offset-(old->offset+old->size);
while(padnum--)
fwrite(instPad,1,1,fp);//填充
}
old=b;
fwrite(b->data,b->size,1,fp);
}
}
}
fclose(fp);
exe.writeElf(dir,2);//输出链接后的elf后半段
}

// Link 执行链接过程
func (l *Linker) Link(outputFile string) error {
	if err := l.CollectInfo(); err != nil {
		return fmt.Errorf("failed to collect info: %v", err)
	}

	if !l.SymValid() {
		return fmt.Errorf("symbol validation failed")
	}

	if err := l.AllocAddr(); err != nil {
		return fmt.Errorf("failed to allocate addresses: %v", err)
	}

	if err := l.SymParser(); err != nil {
		return fmt.Errorf("failed to parse symbols: %v", err)
	}

	if err := l.Relocate(); err != nil {
		return fmt.Errorf("failed to relocate symbols: %v", err)
	}

	return nil
}
