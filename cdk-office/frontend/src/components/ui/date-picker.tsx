import * as React from "react"
import { format, isValid, parse } from "date-fns"
import { Calendar as CalendarIcon } from "lucide-react"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Calendar } from "@/components/ui/calendar"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { zhCN } from "date-fns/locale"

interface DatePickerProps {
  value?: Date
  onChange?: (date: Date | undefined) => void
  placeholder?: string
  disabled?: boolean
  className?: string
  format?: string
}

const DatePicker = React.forwardRef<HTMLButtonElement, DatePickerProps>(
  ({ value, onChange, placeholder = "选择日期", disabled, className, format: dateFormat = "yyyy-MM-dd", ...props }, ref) => {
    const [open, setOpen] = React.useState(false)

    const formatDate = (date: Date | undefined) => {
      if (!date || !isValid(date)) return ""
      return format(date, dateFormat, { locale: zhCN })
    }

    const handleSelect = (date: Date | undefined) => {
      onChange?.(date)
      setOpen(false)
    }

    return (
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            ref={ref}
            variant="outline"
            className={cn(
              "w-full justify-start text-left font-normal",
              !value && "text-muted-foreground",
              className
            )}
            disabled={disabled}
            {...props}
          >
            <CalendarIcon className="mr-2 h-4 w-4" />
            {value ? formatDate(value) : placeholder}
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto p-0" align="start">
          <Calendar
            mode="single"
            selected={value}
            onSelect={handleSelect}
            disabled={disabled}
            locale={zhCN}
            initialFocus
          />
        </PopoverContent>
      </Popover>
    )
  }
)
DatePicker.displayName = "DatePicker"

interface DateTimePickerProps extends DatePickerProps {
  showTime?: boolean
  timeFormat?: string
}

const DateTimePicker = React.forwardRef<HTMLButtonElement, DateTimePickerProps>(
  ({ 
    value, 
    onChange, 
    placeholder = "选择日期时间", 
    disabled, 
    className,
    showTime = true,
    timeFormat = "HH:mm",
    format: dateFormat = "yyyy-MM-dd HH:mm",
    ...props 
  }, ref) => {
    const [open, setOpen] = React.useState(false)
    const [timeValue, setTimeValue] = React.useState("")

    React.useEffect(() => {
      if (value && isValid(value)) {
        setTimeValue(format(value, timeFormat))
      }
    }, [value, timeFormat])

    const formatDate = (date: Date | undefined) => {
      if (!date || !isValid(date)) return ""
      return format(date, dateFormat, { locale: zhCN })
    }

    const handleDateSelect = (date: Date | undefined) => {
      if (!date) {
        onChange?.(undefined)
        return
      }

      if (timeValue && showTime) {
        try {
          const timeDate = parse(timeValue, timeFormat, new Date())
          if (isValid(timeDate)) {
            date.setHours(timeDate.getHours())
            date.setMinutes(timeDate.getMinutes())
          }
        } catch (error) {
          console.warn("Invalid time format:", timeValue)
        }
      }

      onChange?.(date)
      if (!showTime) {
        setOpen(false)
      }
    }

    const handleTimeChange = (event: React.ChangeEvent<HTMLInputElement>) => {
      const newTimeValue = event.target.value
      setTimeValue(newTimeValue)

      if (value && newTimeValue) {
        try {
          const timeDate = parse(newTimeValue, timeFormat, new Date())
          if (isValid(timeDate)) {
            const newDate = new Date(value)
            newDate.setHours(timeDate.getHours())
            newDate.setMinutes(timeDate.getMinutes())
            onChange?.(newDate)
          }
        } catch (error) {
          console.warn("Invalid time format:", newTimeValue)
        }
      }
    }

    return (
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            ref={ref}
            variant="outline"
            className={cn(
              "w-full justify-start text-left font-normal",
              !value && "text-muted-foreground",
              className
            )}
            disabled={disabled}
            {...props}
          >
            <CalendarIcon className="mr-2 h-4 w-4" />
            {value ? formatDate(value) : placeholder}
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto p-0" align="start">
          <Calendar
            mode="single"
            selected={value}
            onSelect={handleDateSelect}
            disabled={disabled}
            locale={zhCN}
            initialFocus
          />
          {showTime && (
            <div className="border-t p-3">
              <div className="flex items-center space-x-2">
                <label className="text-sm font-medium">时间:</label>
                <input
                  type="time"
                  value={timeValue}
                  onChange={handleTimeChange}
                  className="flex h-9 rounded-md border border-input bg-background px-3 py-1 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                />
              </div>
              <div className="mt-2 flex justify-end">
                <Button size="sm" onClick={() => setOpen(false)}>
                  确认
                </Button>
              </div>
            </div>
          )}
        </PopoverContent>
      </Popover>
    )
  }
)
DateTimePicker.displayName = "DateTimePicker"

export { DatePicker, DateTimePicker }