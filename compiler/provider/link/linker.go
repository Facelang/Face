package link

import (
	"face-lang/compiler/target/elf"
	"fmt"
)

const (
	StartSymbol = "_start"   // 程序入口符号
	BaseAddr    = 0x08040000 // 默认加载地址
)

// SymLink 表示符号链接， 符号引用对象，每一次外部符号引用都会产生
type SymLink struct {
	Name string
	Recv *elf.File // 引用符号的文件
	Prov *elf.File // 提供符号的文件， 符号未定义时必然是NULL
}

// Linker 是链接器的主要结构
type Linker struct {
	Exe        *elf.File               // 链接后的输出文件
	StartOwner *elf.File               // 拥有全局符号START/_start的文件
	Elfs       []*elf.File             // 所有目标文件对象
	SymLinks   []*SymLink              // 所有符号引用信息，符号解析前存储未定义的符号prov字段为NULL
	SymDef     []*SymLink              // 所有符号定义信息recv字段NULL时标示该符号没有被任何文件引用，否则指向本身（同prov）
	SegNames   []string                // 链接关心的段
	SegLists   map[string]*elf.ProgSeg // 所有合并段表序列
}

// NewLinker 创建新的链接器
func NewLinker() *Linker {
	l := &Linker{
		SegNames: []string{".text", ".data", ".bss"}, // 初始化三个段, 确保顺序，前三个段分别是...
		SegLists: make(map[string]*elf.ProgSeg),
		SymLinks: make([]*SymLink, 0),
		SymDef:   make([]*SymLink, 0),
	}
	// 初始化段列表
	for _, name := range l.SegNames {
		l.SegLists[name] = &elf.ProgSeg{
			OwnerList: make([]*elf.File, 0),
			Blocks:    make([]*elf.Block, 0),
		}
	}
	return l
}

// AddElf 添加一个目标文件
func (l *Linker) AddElf(name string) error {
	file, err := elf.ReadElf(name)
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
	for _, e := range l.Elfs {
		for _, seg := range l.SegNames { // 初始 .text .data .bss 就三个段
			if _, ok := e.ShdrTab[seg]; ok {
				// seglist 记录具有该段的 elf 文件
				l.SegLists[seg].OwnerList = append(l.SegLists[seg].OwnerList, e)
			}
		}

		for _, name := range e.SymNames { // 每一个文件的符号
			if name == "" {
				continue
			}
			sym := e.SymTab[name]
			link := SymLink{Name: name}
			// todo 还可以判断全局符号，这里直接省略
			if sym.Shndx == uint16(elf.SHN_UNDEF) { // 未定义的符号
				link.Recv = e
				l.SymLinks = append(l.SymLinks, &link) // 需要从其他文件查找符号
			} else { // 已定义的符号
				link.Prov = e
				l.SymDef = append(l.SymDef, &link) // 提供给其它文件
			}
		}
	}
	return nil
}

// SymValid 验证符号： 查找入口文件，校验重复定义的全局变量， 校验未定义的符号
func (l *Linker) SymValid() bool {
	flag := true
	l.StartOwner = nil // 入口函数所在的elf文件
	// 检查定义符号，找入口和重定义
	for i, def := range l.SymDef { // 所有符号提供者 prov
		//if(ELF32_ST_BIND(symDef[i]->prov->symTab[symDef[i]->name]->st_info)!=STB_GLOBAL)//只考虑全局符号
		//continue;
		if def.Name == StartSymbol { // 记录程序入口文件
			l.StartOwner = def.Prov
		}
		// 校验符号是否重复定义， 这里应该校验全局符号
		if def.Prov.SymTab[def.Name].Info>>4 != 1 {
			continue
		}
		for j := i + 1; j < len(l.SymDef); j++ {
			jDef := l.SymDef[j]
			if jDef.Prov.SymTab[jDef.Name].Info>>4 != 1 {
				continue
			}
			if def.Name == jDef.Name { //同名符号
				fmt.Printf("符号名%s在文件%s和文件%s中发生链接冲突。\n", def.Name, def.Prov.Name, jDef.Prov.Name)
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
		if link.Prov == nil { // 未定义 todo 这里发现 main 函数未定义
			info := link.Recv.SymTab[link.Name].Info
			iType := "符号"
			if info&0xf == 1 {
				iType = "变量"
			} else if info&0xf == 2 {
				iType = "函数"
			}
			fmt.Printf("文件%s的%s名%s未定义。\n", link.Recv.Name, iType, link.Name)
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
func (l *Linker) Relocate() error {
	// 重定位项符号必然在符号表中，且地址已经解析完毕
	for _, e := range l.Elfs {
		for _, rel := range e.RelTab {
			sym := e.SymTab[rel.RelName]                            //重定位符号信息
			symAddr := sym.Value                                    //解析后的符号段偏移为虚拟地址
			relAddr := e.ShdrTab[rel.SegName].Addr + rel.Rel.Offset //重定位地址
			l.SegLists[rel.SegName].RelocAddr(relAddr, uint8(rel.Rel.Info&0xff), symAddr)
		}
	}
	return nil
}

// ExportElf 文件组装 生成 ELF 可执行文件
func (l *Linker) ExportElf() *elf.File {
	magic := elf.Elf_Magic{0x7f, 'E', 'L', 'F', 0x1, 0x1, 0x1}
	target := elf.NewElfFile(magic, elf.Elf32_Half(elf.ET_EXEC), elf.Elf32_Half(elf.EM_386))

	// 数据位置指针, 32 怎么来的
	curOff := uint32(target.Ehdr.Ehsize) + uint32(32*len(l.SegNames)) // 文件头52B+程序头表项32*个数

	// 计算节名字符串表大小， 书中直接将三个名字 和 l.SegNames 添加到 一个 AllSegNames, 然后添加到 shstrtab(shindex 根据名字记录索引，shstrindex 记录名字偏移)
	shstrtabSize := 26 // ".shstrtab".length()+".symtab".length()+".strtab".length()+3(空格)
	for _, name := range l.SegNames {
		seg := l.SegLists[name]       // 合并后的段
		shstrtabSize += len(name) + 1 // 考虑结束符'\0'
		target.AddProgSeg(name, seg)  // 添加头表会同时添加段表

		//计算有效数据段的大小和偏移,最后一个决定
		curOff = seg.Offset //修正当前偏移，循环结束后保留的是.bss的基址
	}

	target.Ehdr.Phoff = 52
	target.Ehdr.Phentsize = 32 // 程序头表大小 不包含新添加的三个段
	target.Ehdr.Phnum = elf.Elf32_Half(len(l.SegNames))
	target.Shstrtab = make([]byte, shstrtabSize)
	target.ShstrtabSize = shstrtabSize

	// todo 可优化，段表字符串信息，可以在添加段表时生成
	index := 0
	shstrIndex := make(map[string]int) //段表串名与索引映射
	shstrIndex[".shstrtab"] = index
	copy(target.Shstrtab[index:], ".shstrtab")
	index += 10

	shstrIndex[".symtab"] = index
	copy(target.Shstrtab[index:], ".symtab")
	index += 8

	shstrIndex[".strtab"] = index
	copy(target.Shstrtab[index:], ".strtab")
	index += 8

	shstrIndex[""] = index - 1

	for _, name := range l.SegNames {
		shstrIndex[name] = index
		copy(target.Shstrtab[index:], name)
		index += len(name) + 1 // 留一个 \x0
	}

	// .shstrtab
	target.AddShdr(".shstrtab", elf.Elf32_Word(elf.SHT_STRTAB), 0, 0,
		curOff, elf.Elf32_Word(shstrtabSize), elf.Elf32_Word(elf.SHN_UNDEF), 0, 1, 0)
	target.Ehdr.Shstrndx = elf.Elf32_Half(target.GetSymIndex(".shstrtab"))
	curOff += uint32(shstrtabSize) //段表偏移
	target.Ehdr.Shoff = curOff
	target.Ehdr.Shentsize = 40
	target.Ehdr.Shnum = elf.Elf32_Half(4 + len(l.SegNames))

	//生成符号表项
	curOff += 40 * (4 + uint32(target.Ehdr.Shnum)) //符号表偏移
	target.AddShdr(".symtab", elf.Elf32_Word(elf.SHT_SYMTAB), 0, 0,
		curOff, elf.Elf32_Word((1+len(l.SymDef))*16), 0, 0, 1, 16)
	target.ShdrTab[".symtab"].Link = elf.Elf32_Word(target.GetSegIndex(".symtab") + 1) //。strtab默认在.symtab之后
	strtabSize := 0                                                                    //字符串表大小
	for _, link := range l.SymDef {                                                    //遍历所有符号
		strtabSize += len(link.Name) + 1
		sym := link.Prov.SymTab[link.Name]
		sym.Shndx = uint16(target.GetSegIndex(link.Prov.ShdrNames[sym.Shndx]))
		target.AddSym(link.Name, sym)
	}

	// 设置程序入口点
	target.Ehdr.Entry = target.SymTab[StartSymbol].Value

	// .strtab偏移
	curOff += uint32((1 + len(l.SymDef)) * 16)
	// 添加 .strtab
	target.AddShdr(".strtab", elf.Elf32_Word(elf.SHT_STRTAB), 0, 0,
		curOff, uint32(strtabSize), 0, 0, 1, 0)

	// 填充字符串表
	target.Strtab = make([]byte, strtabSize)
	target.StrtabSize = strtabSize

	//串表与索引映射
	index = 0
	strIndex := make(map[string]int)
	strIndex[""] = strtabSize - 1
	for _, link := range l.SymDef {
		strIndex[link.Name] = index
		name := fmt.Sprintf("%s\x00", link.Name)
		copy(target.Strtab[index:], name)
		index += len(link.Name) + 1
	}

	//更新符号表name
	for n, sym := range target.SymTab {
		sym.Name = uint32(strIndex[n])
	}

	//更新段表name
	for n, shdr := range target.ShdrTab {
		shdr.Name = uint32(strIndex[n])
	}

	return target

}
