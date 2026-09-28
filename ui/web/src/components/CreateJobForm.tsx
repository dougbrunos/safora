import { useState } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import * as z from "zod"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"

const jobSchema = z.object({
  name: z.string().min(2, "Name must be at least 2 characters"),
  source: z.string().min(1, "Source is required"),
  destination: z.string().min(1, "Destination is required"),
  enableVSS: z.boolean().default(false),
  retention: z.string().default("KEEP 5"),
  exclusions: z.string().optional(),
})



interface Props {
  onSuccess: () => void;
  onCancel: () => void;
  jobToEdit?: any;
}

export function CreateJobForm({ onSuccess, onCancel, jobToEdit }: Props) {
  const [loading, setLoading] = useState(false)

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    formState: { errors },
  } = useForm({
    resolver: zodResolver(jobSchema),
    defaultValues: {
      name: jobToEdit?.Name || "",
      source: jobToEdit?.Sources?.[0]?.Path || "",
      destination: jobToEdit?.Destinations?.[0]?.Path || "",
      enableVSS: jobToEdit?.Description === "VSS Enabled" || false,
      retention: jobToEdit?.RetentionPolicy || "KEEP 5",
      exclusions: jobToEdit?.Sources?.[0]?.ExclusionRules || "",
    },
  })

  const enableVSS = watch("enableVSS")
  const retention = watch("retention")

  const onSubmit = async (data: any) => {
    setLoading(true)
    try {
      const payload = {
        Name: data.name,
        StorageStrategy: jobToEdit?.StorageStrategy || "Date-Stamped Mirroring",
        RetentionPolicy: data.retention,
        Description: data.enableVSS ? "VSS Enabled" : "",
        RetryCount: jobToEdit?.RetryCount || 3,
        RetryWait: jobToEdit?.RetryWait || 30,
        LogOutput: jobToEdit?.LogOutput || "",
        Sources: [{ Path: data.source, ExclusionRules: data.exclusions || "" }],
        Destinations: [{ Path: data.destination }],
      }

      const url = jobToEdit ? `/api/jobs/${jobToEdit.ID}` : "/api/jobs"
      const method = jobToEdit ? "PUT" : "POST"

      const res = await fetch(url, {
        method,
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      })

      if (res.ok) {
        onSuccess()
      } else {
        alert("Failed to create job")
      }
    } catch (e) {
      console.error(e)
      alert("Error creating job")
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      <div className="space-y-2">
        <Label htmlFor="name">Job Name</Label>
        <Input id="name" placeholder="e.g. Daily Backup" {...register("name")} />
        {errors.name && <p className="text-sm text-destructive">{errors.name.message}</p>}
      </div>

      <div className="space-y-2">
        <Label htmlFor="source">Source Path</Label>
        <Input id="source" placeholder="C:\Data\{yesterday}" {...register("source")} />
        <p className="text-xs text-muted-foreground">
          Safora supports dynamic templates like <code>{'{yesterday}'}</code> and <code>{'{today}'}</code>.
        </p>
        {errors.source && <p className="text-sm text-destructive">{errors.source.message}</p>}
      </div>

      <div className="space-y-2">
        <Label htmlFor="destination">Destination Path</Label>
        <Input id="destination" placeholder="D:\Backup" {...register("destination")} />
        {errors.destination && <p className="text-sm text-destructive">{errors.destination.message}</p>}
      </div>

      <div className="space-y-2">
        <Label htmlFor="exclusions">Exclusions (optional)</Label>
        <Input id="exclusions" placeholder="DIR:temp,FILE:*.log" {...register("exclusions")} />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Retention Policy</Label>
          <Select value={retention as string} onValueChange={(val) => setValue("retention", val as string)}>
            <SelectTrigger>
              <SelectValue placeholder="Select retention" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="KEEP 3">Keep last 3</SelectItem>
              <SelectItem value="KEEP 5">Keep last 5</SelectItem>
              <SelectItem value="KEEP 10">Keep last 10</SelectItem>
              <SelectItem value="KEEP 30">Keep last 30</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div className="flex flex-col justify-center space-y-2 pt-2">
          <div className="flex items-center space-x-2">
            <Switch
              id="vss"
              checked={!!enableVSS}
              onCheckedChange={(val) => setValue("enableVSS", val)}
            />
            <Label htmlFor="vss">Enable VSS</Label>
          </div>
          <p className="text-xs text-muted-foreground">Snapshot locked files</p>
        </div>
      </div>

      <div className="flex justify-end gap-3 pt-4 border-t border-border">
        <Button type="button" variant="ghost" onClick={onCancel}>Cancel</Button>
        <Button type="submit" disabled={loading}>
          {loading ? "Saving..." : jobToEdit ? "Save Changes" : "Create Job"}
        </Button>
      </div>
    </form>
  )
}
