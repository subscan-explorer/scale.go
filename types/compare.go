package types

import (
	"fmt"
	"strings"
)

// Compare compare metadata
// dest is older metadata

type MetadataCompareResult struct {
	NewModules    []string                 `json:"new_modules,omitempty"`
	ModuleChanges map[string]ModuleChanges `json:"module_changes"`
}

type ModuleChanges struct {
	Calls   *ChangesList `json:"calls,omitempty"`
	Events  *ChangesList `json:"events,omitempty"`
	Storage *ChangesList `json:"storage,omitempty"`
}

type ChangesList struct {
	New     []string         `json:"new,omitempty"`
	Changes []CompareChanges `json:"changes,omitempty"`
}

type CompareChanges struct {
	Prev    string `json:"prev"`
	Current string `json:"next"`
}

func (m *MetadataTag) Compare(dest *MetadataTag) *MetadataCompareResult {
	var getModuleByName = func(name string) *MetadataModules {
		for _, module := range dest.Modules {
			if module.Name == name {
				return &module
			}
		}
		return nil
	}

	var getCallByName = func(name string, calls []MetadataCalls) *MetadataCalls {
		for _, call := range calls {
			if call.Name == name {
				return &call
			}
		}
		return nil
	}

	var getEventByName = func(name string, events []MetadataEvents) *MetadataEvents {
		for _, event := range events {
			if event.Name == name {
				return &event
			}
		}
		return nil
	}

	var getStorageByName = func(name string, storageItems []MetadataStorage) *MetadataStorage {
		for _, item := range storageItems {
			if item.Name == name {
				return &item
			}
		}
		return nil
	}

	var buildCallArgs = func(call MetadataCalls) string {
		var args []string
		for _, arg := range call.Args {
			args = append(args, arg.Type)
		}
		return fmt.Sprintf("%s(%s)", call.Name, strings.Join(args, ","))
	}

	var buildEventArgs = func(event MetadataEvents) string {
		return fmt.Sprintf("%s(%s)", event.Name, strings.Join(event.Args, ","))
	}

	var buildStorageArgs = func(moduleName string, storage MetadataStorage) string {
		return fmt.Sprintf("%s.%s: %s", moduleName, storage.Name, storage.Type.TypeValue())
	}

	var result MetadataCompareResult
	result.ModuleChanges = make(map[string]ModuleChanges)
	for _, module := range m.Modules {
		destModule := getModuleByName(module.Name)
		if destModule == nil {
			// new modules
			result.NewModules = append(result.NewModules, module.Name)
			continue
		}

		// check call
		moduleChanges := ModuleChanges{}
		calls := ChangesList{}
		for _, call := range module.Calls {
			destCall := getCallByName(call.Name, destModule.Calls)
			if destCall == nil {
				calls.New = append(calls.New, buildCallArgs(call))
				continue
			}
			// check call args
			if len(call.Args) != len(destCall.Args) {
				calls.Changes = append(calls.Changes, CompareChanges{
					Prev:    buildCallArgs(*destCall),
					Current: buildCallArgs(call),
				})
				continue
			}
			for index, arg := range call.Args {
				// check every type
				destType := destCall.Args[index].Type
				if !Eq(arg.Type, destType) {
					calls.Changes = append(calls.Changes, CompareChanges{
						Prev:    buildCallArgs(*destCall),
						Current: buildCallArgs(call),
					})
					continue
				}
			}
		}
		if calls.New != nil || calls.Changes != nil {
			moduleChanges.Calls = &calls
		}

		// check event
		Events := ChangesList{}
		for _, event := range module.Events {
			destEvent := getEventByName(event.Name, destModule.Events)
			if destEvent == nil {
				Events.New = append(Events.New, buildEventArgs(event))
				continue
			}
			// check call args
			if len(event.Args) != len(destEvent.Args) {
				Events.Changes = append(Events.Changes, CompareChanges{
					Prev:    buildEventArgs(*destEvent),
					Current: buildEventArgs(event),
				})
				continue
			}
			for index, arg := range event.Args {
				// check every type
				if !Eq(arg, destEvent.Args[index]) {
					Events.Changes = append(Events.Changes, CompareChanges{
						Prev:    buildEventArgs(*destEvent),
						Current: buildEventArgs(event),
					})
					continue
				}
			}
		}
		if Events.New != nil || Events.Changes != nil {
			moduleChanges.Events = &Events
		}

		// check storage
		storage := ChangesList{}
		for _, storageItem := range module.Storage {
			destStorage := getStorageByName(storageItem.Name, destModule.Storage)
			if destStorage == nil {
				storage.New = append(storage.New, fmt.Sprintf("%s.%s", module.Name, storageItem.Name))
				continue
			}

			if !Eq(storageItem.Type.TypeValue(), destStorage.Type.TypeValue()) {
				storage.Changes = append(storage.Changes, CompareChanges{
					Prev:    buildStorageArgs(module.Name, *destStorage),
					Current: buildStorageArgs(module.Name, storageItem),
				})
			}
		}
		if storage.New != nil || storage.Changes != nil {
			moduleChanges.Storage = &storage
		}

		if (moduleChanges.Calls != nil && (len(moduleChanges.Calls.Changes) > 0 || len(moduleChanges.Calls.New) > 0)) ||
			(moduleChanges.Events != nil && (len(moduleChanges.Events.Changes) > 0 || len(moduleChanges.Events.New) > 0)) ||
			(moduleChanges.Storage != nil && (len(moduleChanges.Storage.Changes) > 0 || len(moduleChanges.Storage.New) > 0)) {
			result.ModuleChanges[module.Name] = moduleChanges
		}

	}
	return &result
}
