import * as React from "react"
import { Calendar as CalendarIcon } from "lucide-react"
import type { DateRange } from "react-day-picker"
import { format } from "date-fns"

import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Calendar } from "@/components/ui/calendar"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"

interface DatePickerWithRangeProps {
  className?: string
  onDateRangeChange?: (range: DateRange | undefined) => void
  dateRange?: DateRange
}

export function DatePickerWithRange({
  className,
  onDateRangeChange,
  dateRange
}: DatePickerWithRangeProps) {
  const [date, setDate] = React.useState<DateRange | undefined>(dateRange)

  // Update internal state when prop changes
  React.useEffect(() => {
    setDate(dateRange)
  }, [dateRange])

  const handleDateChange = (range: DateRange | undefined) => {
    setDate(range)
    onDateRangeChange?.(range)
  }

  return (
    <div className={cn("grid gap-2", className)}>
      <Popover>
        <PopoverTrigger asChild>
          <Button
            id="date"
            variant={"outline"}
            className={cn(
              "w-[260px] justify-start text-left font-normal",
              !date && "text-muted-foreground"
            )}
          >
            <CalendarIcon className="mr-2 h-4 w-4" />
            {date?.from ? (
              date.to ? (
                <>
                  {format(date.from, "MMMM dd, yyyy")} -{" "}
                  {format(date.to, "MMMM dd, yyyy")}
                </>
              ) : (
                format(date.from, "MMMM dd, yyyy")
              )
            ) : (
              <span>Pick a date range</span>
            )}
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto p-0" align="center" side="bottom" sideOffset={60}>
          <div className="p-3">
            <Calendar
              initialFocus
              mode="range"
              defaultMonth={date?.from}
              selected={date}
              onSelect={handleDateChange}
              numberOfMonths={2}
              
            />
            <div className="flex justify-end space-x-2 mt-2">
              <Button variant="outline" size="sm" onClick={() => handleDateChange(undefined)}>
                Clear
              </Button>
            </div>
          </div>
        </PopoverContent>
      </Popover>
    </div>
  )
}