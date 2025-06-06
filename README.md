# Facelang

Facelang 是一个全新编程语言项目。该项目期望打造一门专为全栈工程师量身定制的跨端编译语言，目前主要基于 Go 实现。

下一阶段，会优先使用 LLVM 完成语言的整体功能，确保语言基本可用。

后期仍然计划参考 Go 实现更独立，更完善的汇编器和链接器，实现轻量化的编译。

该项目目前属于个人维护，大部分功能实现完成度不高，开源的目的是希望对此项目感兴趣的朋友可以一起加入学习、探讨。

项目中可能会出现诸多问题，还望大家多多包涵。

🤘 🤘 🤘 自嗨 ing ...

致歉。

## 🚪 快速上手

- COME SOON ...

## 🚀 项目特性

- 完整的编译器实现
- 汇编器支持
- 链接器功能
- 跨平台支持
- 跨端GUI支持
- 底层原理学习

## 🛠️ 技术栈

- Go 语言
- x86 汇编
- Arm 汇编
- llvm
- C/C++

## 📚 项目结构

```
face-lang/
├── compiler/     # 编译器实现
├── docs/         # 文档
├── example/      # 示例代码
├── internal/     # 核心代码
├── library/      # 标准库
└── tools/        # 反汇编工具
```

## 🎯 项目目标

1. 实现一门完整的编程语言
2. 提供跨平台编译支持
3. 为全栈工程师提供高效的开发体验
4. 探索和学习程序运行的底层原理

## 🏃‍♂‍➡ 开发进度

- [x] Linux 平台支持
- [ ] Osx 平台支持
- [ ] Windows 平台支持
- [x] 基础汇编指令支持 `mov`、`cmp`、`sub`、`add`、`lea`、`call`、`int`、`imul`、`idiv`、`neg`、`inc`、`dec`、`jmp`、`je`、`jg`、`jl`、`jle`、`jne`、`jna`、`push`、`pop`
- [ ] 基于 LLVM 实现
- [ ] 文档完善
- [ ] 其它汇编指令支持
- [ ] 标准库完善
- [ ] 性能优化

## 🤝 参与贡献

欢迎所有对编程语言实现感兴趣的朋友参与项目开发！您可以通过以下方式参与：

1. 提交 Issue 报告问题或建议
2. 提交 Pull Request 贡献代码
3. 完善项目文档
4. 分享使用经验

## 📝 项目笔记

TODO

## 🔮 未来规划

- [ ] 实现更多语言特性
- [ ] 优化编译性能
- [ ] 提供更多平台支持
- [ ] 完善开发工具链
- [ ] 建立活跃的社区

## 📄 开源协议

本项目采用 [Creative Commons Attribution 4.0 International License](LICENSE) 协议开源。

根据该协议，您可以：
- 自由分享和分发本项目
- 自由修改和构建本项目
- 用于任何目的，包括商业用途

主要要求：
- 必须注明原作者
- 必须提供许可证链接
- 必须说明是否做了修改

## 🌟 致谢

感谢所有为项目做出贡献的开发者！

---

欢迎关注项目，一起探索编程语言的奥秘！ 