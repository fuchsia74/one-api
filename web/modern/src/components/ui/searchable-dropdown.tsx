import * as React from 'react'
import { Check, ChevronsUpDown, Loader2, Plus } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'

export interface SearchOption {
  key: string
  value: string
  text: string
  content?: React.ReactNode
}

interface SearchableDropdownProps {
  value?: string
  placeholder?: string
  searchPlaceholder?: string
  options: SearchOption[]
  onSearchChange?: (query: string) => void
  onChange?: (value: string) => void
  onAddItem?: (value: string) => void
  loading?: boolean
  noResultsMessage?: string
  additionLabel?: string
  allowAdditions?: boolean
  clearable?: boolean
  className?: string
}

export function SearchableDropdown({
  value,
  placeholder = "Select option...",
  searchPlaceholder = "Search...",
  options,
  onSearchChange,
  onChange,
  onAddItem,
  loading = false,
  noResultsMessage = "No results found",
  additionLabel = "Add: ",
  allowAdditions = false,
  clearable = false,
  className,
}: SearchableDropdownProps) {
  const [open, setOpen] = React.useState(false)
  const [searchValue, setSearchValue] = React.useState('')

  const selectedOption = options.find((option) => option.value === value)

  const handleSearchChange = (query: string) => {
    setSearchValue(query)
    onSearchChange?.(query)
  }

  const handleSelect = (selectedValue: string) => {
    if (selectedValue === value && clearable) {
      onChange?.('')
    } else {
      onChange?.(selectedValue)
    }
    setOpen(false)
  }

  const handleAddItem = () => {
    if (searchValue && allowAdditions && onAddItem) {
      onAddItem(searchValue)
      onChange?.(searchValue)
      setSearchValue('')
      setOpen(false)
    }
  }

  const filteredOptions = React.useMemo(() => {
    if (!searchValue) return options
    return options.filter((option) =>
      option.text.toLowerCase().includes(searchValue.toLowerCase()) ||
      option.value.toLowerCase().includes(searchValue.toLowerCase())
    )
  }, [options, searchValue])

  const showAddition = allowAdditions &&
    searchValue &&
    !filteredOptions.some(option =>
      option.value.toLowerCase() === searchValue.toLowerCase()
    )

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className={cn("w-full justify-between", className)}
        >
          {selectedOption ? (
            <span className="truncate">{selectedOption.text}</span>
          ) : (
            <span className="text-muted-foreground">{placeholder}</span>
          )}
          <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-full p-0" align="start">
        <Command>
          <CommandInput
            placeholder={searchPlaceholder}
            value={searchValue}
            onValueChange={handleSearchChange}
          />
          <CommandList>
            {loading ? (
              <div className="flex items-center justify-center py-6">
                <Loader2 className="h-4 w-4 animate-spin" />
                <span className="ml-2 text-sm text-muted-foreground">Loading...</span>
              </div>
            ) : (
              <>
                {filteredOptions.length === 0 && !showAddition ? (
                  <CommandEmpty>{noResultsMessage}</CommandEmpty>
                ) : (
                  <CommandGroup>
                    {filteredOptions.map((option) => (
                      <CommandItem
                        key={option.key}
                        value={option.value}
                        onSelect={handleSelect}
                      >
                        <Check
                          className={cn(
                            "mr-2 h-4 w-4",
                            value === option.value ? "opacity-100" : "opacity-0"
                          )}
                        />
                        {option.content || option.text}
                      </CommandItem>
                    ))}
                    {showAddition && (
                      <CommandItem onSelect={handleAddItem}>
                        <Plus className="mr-2 h-4 w-4" />
                        {additionLabel}{searchValue}
                      </CommandItem>
                    )}
                  </CommandGroup>
                )}
              </>
            )}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
