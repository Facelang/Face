package link

import (
	"fmt"
)

func Link(target string, includes ...string) error {
	linker := NewLinker()
	for _, include := range includes {
		err := linker.AddElf(include)
		if err != nil {
			return err
		}
	}

	if err := linker.CollectInfo(); err != nil {
		return fmt.Errorf("failed to collect info: %v", err)
	}

	if !linker.SymValid() {
		return fmt.Errorf("symbol validation failed")
	}

	if err := linker.AllocAddr(); err != nil {
		return fmt.Errorf("failed to allocate addresses: %v", err)
	}

	if err := linker.SymParser(); err != nil {
		return fmt.Errorf("failed to parse symbols: %v", err)
	}

	if err := linker.Relocate(); err != nil {
		return fmt.Errorf("failed to relocate symbols: %v", err)
	}

	//if err := linker.AssemExe(writer); err != nil {
	//	return fmt.Errorf("failed to relocate symbols: %v", err)
	//}

	export := linker.ExportElf()
	return export.WriteFile(target)

}
