package engine

type indexed struct {
	added   []compiled
	deleted []compiled
	any     []compiled
}

func buildIndex(items []compiled) indexed {
	index := indexed{}
	for _, item := range items {
		if item.rule.Target == "added" {
			index.added = append(index.added, item)
			continue
		}
		if item.rule.Target == "deleted" {
			index.deleted = append(index.deleted, item)
			continue
		}
		index.any = append(index.any, item)
	}
	return index
}

func (index indexed) selectRules(kind string) []compiled {
	items := make([]compiled, 0)
	if kind == "added" {
		items = append(items, index.added...)
	}
	if kind == "deleted" {
		items = append(items, index.deleted...)
	}
	items = append(items, index.any...)
	return items
}
