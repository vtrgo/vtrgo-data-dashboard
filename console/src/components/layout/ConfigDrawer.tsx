// console/src/components/ConfigDrawer.tsx
import {
  Drawer,
  DrawerContent,
  DrawerHeader,
  DrawerTitle,
  DrawerDescription,
  DrawerClose,
} from "@/components/ui/drawer"
import { Button } from "@/components/ui/button"

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function ConfigDrawer({ open, onOpenChange }: Props) {
  return (
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DrawerContent className="p-6 max-w-md ml-auto bg-[url('/textures/paper-fiber.png')] bg-repeat border-l shadow-lg">
        <DrawerHeader>
          <DrawerTitle>Configuration</DrawerTitle>
          <DrawerDescription>
            View and modify system settings.
          </DrawerDescription>
        </DrawerHeader>

        {/* Replace with real config */}
        <div className="space-y-4 py-4">
          <div className="text-sm">
            <strong>PLC Source:</strong> ethernet-ip
          </div>
          <div className="text-sm">
            <strong>Poll Interval:</strong> 1000 ms
          </div>
          <div className="text-sm">
            <strong>Influx Bucket:</strong> status_data
          </div>
        </div>

        <DrawerClose asChild>
          <Button variant="outline" className="mt-4">Close</Button>
        </DrawerClose>
      </DrawerContent>
    </Drawer>
  )
}
