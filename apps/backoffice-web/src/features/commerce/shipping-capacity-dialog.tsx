import { useState } from "react";
import {
  Alert,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  LinearProgress,
  TextField,
} from "@mui/material";
import { useQuery } from "@tanstack/react-query";
import { useWorkspace } from "../../shared/workspace";
import { ApiError } from "../../shared/api";
import { DataTable } from "../../shared/components";
import type { ShippingAvailability } from "../../../../../packages/commerce-client/types";
import {
  parsePersianCapacityDate,
  persianCapacityDate,
  tehranCapacityDate,
} from "./shipping-dates";

type Capacity = {
  preview?: ShippingAvailability["methods"][number];
  store_id: string;
  method_id: string;
  date: string;
  active: boolean;
  open: boolean;
  limit: number | null;
  used: number;
  held: number;
  committed: number;
  remaining: number;
  slots: {
    window: {
      window_id: string;
      start_at: string;
      end_at: string;
      capacity: number;
    };
    used: number;
    remaining: number;
  }[];
};
const clock = (value: string) =>
  new Date(value).toLocaleTimeString("fa-IR", {
    timeZone: "Asia/Tehran",
    hour: "2-digit",
    minute: "2-digit",
  });
const number = (value: number) => value.toLocaleString("fa-IR");
const dateLabel = (value: string) =>
  new Date(value + "T12:00:00+03:30").toLocaleDateString("fa-IR", {
    timeZone: "Asia/Tehran",
    weekday: "long",
    month: "long",
    day: "numeric",
  });

export function ShippingCapacityDialog({
  storeID,
  methodID,
  name,
  close,
}: {
  storeID: string;
  methodID: string;
  name: string;
  close: () => void;
}) {
  const { authToken, surface } = useWorkspace();
  const [date, setDate] = useState(tehranCapacityDate);
  const [dateText, setDateText] = useState(() =>
    persianCapacityDate(tehranCapacityDate()),
  );
  const query = useQuery({
    queryKey: [
      "shipping-capacity",
      surface,
      authToken,
      storeID,
      methodID,
      date,
    ],
    enabled: /^\d{4}-\d{2}-\d{2}$/.test(date),
    refetchOnWindowFocus: false,
    queryFn: async () => {
      const response = await fetch(
        (import.meta.env.VITE_API_BASE_URL ?? "") +
          `/${surface}/stores/shipping/capacity?` +
          new URLSearchParams({ store_id: storeID, method_id: methodID, date }),
        {
          cache: "no-store",
          headers: { Authorization: `Bearer ${authToken}` },
        },
      );
      if (!response.ok) throw new ApiError(response.status);
      return response.json() as Promise<Capacity>;
    },
  });
  const data = query.data;
  return (
    <Dialog open fullWidth maxWidth="md" onClose={close}>
      <DialogTitle>ظرفیت حمل — {name}</DialogTitle>
      <DialogContent>
        <TextField
          label="روز ظرفیت (شمسی)"
          margin="normal"
          value={dateText}
          onChange={(e) => {
            setDateText(e.target.value);
            setDate(parsePersianCapacityDate(e.target.value));
          }}
          placeholder="۱۴۰۵/۰۹/۰۹"
          error={!!dateText && !date}
          inputProps={{ dir: "ltr", maxLength: 10 }}
          InputLabelProps={{ shrink: true }}
          helperText={
            date && Number.isFinite(new Date(date).valueOf())
              ? new Date(date + "T12:00:00+03:30").toLocaleDateString("fa-IR", {
                  dateStyle: "full",
                  timeZone: "Asia/Tehran",
                })
              : "روز را انتخاب کنید"
          }
        />
        <Button
          disabled={query.isFetching || !date}
          onClick={() => void query.refetch()}
        >
          دریافت آخرین مصرف
        </Button>
        {query.isFetching && <LinearProgress aria-label="دریافت ظرفیت" />}
        {query.error && (
          <Alert severity="error">
            دریافت ظرفیت انجام نشد. {query.error.message}
          </Alert>
        )}
        {data && (
          <>
            {data.limit == null && (
              <Alert severity="warning">
                سقف روزانه هنوز تعیین نشده است؛ این روش در خرید تازه ارائه
                نمی‌شود.
              </Alert>
            )}
            {!data.open && (
              <Alert severity="info">
                روش یا روز انتخاب‌شده برای تحویل تازه باز نیست.
              </Alert>
            )}
            {data.limit != null && data.used > data.limit && (
              <Alert severity="warning">
                مصرف قبلی از سقف جدید بیشتر است؛ تعهد قبلی لغو نمی‌شود و ظرفیت
                تازه صفر است.
              </Alert>
            )}
            <div className="commerce-status">
              <div>
                سقف روزانه
                <strong>
                  {data.limit == null ? "تعیین نشده" : number(data.limit)}
                </strong>
              </div>
              <div>
                مصرف کل<strong>{number(data.used)}</strong>
              </div>
              <div>
                رزرو موقت<strong>{number(data.held)}</strong>
              </div>
              <div>
                تثبیت‌شده<strong>{number(data.committed)}</strong>
              </div>
              <div>
                باقیمانده<strong>{number(data.remaining)}</strong>
              </div>
            </div>
            <p>
              هر مرسوله یک واحد؛ ظرفیت پاسخ لحظه‌ای است و هنگام رزرو دوباره
              کنترل می‌شود. روز ظرفیت در روش بدون اسلات، روز ارسال و در روش
              اسلات‌دار، روز تحویل است.
            </p>
            {data.preview && (
              <section className="shipping-capacity-preview">
                <h3>پیش‌نمایش تحویل با فرض پرداخت الان</h3>
                <p>
                  مبنای این نمایش، وقفه و تقویم فروشگاه است؛ روش‌های خرید به
                  آدرس و اقلام مرسوله نیز وابسته‌اند.
                </p>
                {data.preview.range ? (
                  <p>
                    {dateLabel(data.preview.range.delivery_from)} تا{" "}
                    {dateLabel(data.preview.range.delivery_to)}
                  </p>
                ) : (
                  <>
                    <p>
                      از نخستین روز آزاد تا پنج روز تقویمی بعد؛ شش تاریخ ثابت
                    </p>
                    <div
                      className="shipping-preview-days"
                      role="group"
                      aria-label="پیش‌نمایش تاریخ‌های قابل تحویل"
                    >
                      {data.preview.days.map((day) => (
                        <Button
                          key={day.date}
                          disabled={!day.open}
                          aria-pressed={date === day.date}
                          variant={date === day.date ? "contained" : "outlined"}
                          onClick={() => {
                            setDate(day.date);
                            setDateText(persianCapacityDate(day.date));
                          }}
                        >
                          <span>{dateLabel(day.date)}</span>
                          <small>
                            {day.open
                              ? `${number(day.slots.length)} بازهٔ آزاد`
                              : "بسته یا پر"}
                          </small>
                        </Button>
                      ))}
                    </div>
                  </>
                )}
              </section>
            )}
            {!!data.slots.length && (
              <DataTable
                searchable={false}
                rows={data.slots.map((s) => ({ ...s, id: s.window.window_id }))}
                columns={[
                  {
                    title: "بازهٔ تحویل",
                    render: (s) =>
                      `${clock(s.window.start_at)} تا ${clock(s.window.end_at)}`,
                  },
                  {
                    title: "سقف بازه",
                    render: (s) => number(s.window.capacity),
                  },
                  { title: "مصرف", render: (s) => number(s.used) },
                  {
                    title: "باقیمانده با سقف روز",
                    render: (s) => number(s.remaining),
                  },
                ]}
              />
            )}
          </>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={close}>بستن</Button>
      </DialogActions>
    </Dialog>
  );
}
