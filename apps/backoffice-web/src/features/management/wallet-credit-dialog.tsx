import { useEffect, useRef, useState } from "react";
import {
  Alert,
  Button,
  Checkbox,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControlLabel,
  LinearProgress,
  TextField,
} from "@mui/material";
import { useWorkspace } from "../../shared/workspace";
import { ApiError } from "../../shared/api";
import type { User } from "../../shared/types";
import type {
  CustomerCreditWallet,
  WalletCreditInput,
  WalletCreditResult,
} from "../../../../../packages/commerce-client/types";
import { money } from "../../../../../packages/commerce-client/types";

export function WalletCreditDialog({
  user,
  close,
  onCredited,
}: {
  user: User;
  close: () => void;
  onCredited: (result: WalletCreditResult) => void;
}) {
  const { actorID, readCustomerWallet, creditCustomerWallet } = useWorkspace();
  const storageKey = `cherriva:pending-admin-credit:${actorID}:${user.id}`;
  const [recovered] = useState<
    { key: string; input: WalletCreditInput } | undefined
  >(() => {
    try {
      const value = JSON.parse(sessionStorage.getItem(storageKey) ?? "null");
      if (
        value &&
        typeof value.key === "string" &&
        value.input?.user_id === user.id &&
        typeof value.input.amount_toman === "string" &&
        typeof value.input.title === "string" &&
        typeof value.input.description === "string" &&
        typeof value.input.withdrawable === "boolean"
      )
        return value;
    } catch {
      /* Invalid local drafts cannot authorize a financial write. */
    }
    return undefined;
  });
  const key = useRef(recovered?.key ?? crypto.randomUUID());
  const [wallet, setWallet] = useState<CustomerCreditWallet>(),
    [result, setResult] = useState<WalletCreditResult>(),
    [error, setError] = useState(""),
    [busy, setBusy] = useState(true);
  const [form, setForm] = useState<WalletCreditInput>(
    recovered?.input ?? {
      user_id: user.id,
      amount_toman: "",
      title: "",
      description: "",
      withdrawable: false,
    },
  );
  const running = useRef(false),
    submitted = useRef<WalletCreditInput | undefined>(recovered?.input);
  useEffect(() => {
    let active = true;
    readCustomerWallet(user.id, key.current)
      .then((data) => {
        if (!active) return;
        setWallet(data);
        if (data.previous_operation) {
          setResult(data.previous_operation);
          onCredited(data.previous_operation);
          sessionStorage.removeItem(storageKey);
        }
      })
      .catch((e) => {
        if (active) setError(e.message);
      })
      .finally(() => {
        if (active) setBusy(false);
      });
    return () => {
      active = false;
    };
  }, [user.id]);
  const apply = async () => {
    if (running.current || result) return;
    running.current = true;
    setBusy(true);
    setError("");
    const input = submitted.current ?? { ...form };
    submitted.current = input;
    sessionStorage.setItem(
      storageKey,
      JSON.stringify({ key: key.current, input }),
    );
    try {
      const r = await creditCustomerWallet(input, key.current);
      setResult(r);
      onCredited(r);
      sessionStorage.removeItem(storageKey);
    } catch (e) {
      // A lost reply is retried with the exact same key and payload, including after reopen.
      if (e instanceof ApiError && [400, 403, 404].includes(e.status)) {
        submitted.current = undefined;
        sessionStorage.removeItem(storageKey);
      }
      if (e instanceof ApiError && e.status === 409) {
        try {
          const data = await readCustomerWallet(user.id, key.current);
          if (data.previous_operation) {
            setResult(data.previous_operation);
            onCredited(data.previous_operation);
            sessionStorage.removeItem(storageKey);
            return;
          }
        } catch {
          /* Original error stays visible. */
        }
      }
      setError((e as Error).message);
    } finally {
      running.current = false;
      setBusy(false);
    }
  };
  return (
    <Dialog
      open
      fullWidth
      maxWidth="sm"
      onClose={() => {
        if (!busy) close();
      }}
      aria-labelledby="wallet-credit-title"
    >
      <DialogTitle id="wallet-credit-title">
        شارژ کیف پول · {wallet?.customer_name ?? user.name}
      </DialogTitle>
      <DialogContent>
        {busy && <LinearProgress aria-label="دریافت یا ثبت اعتبار" />}
        <p>
          حساب مقصد: <b>{wallet?.customer_name ?? user.name}</b>
          <small
            className="muted"
            style={{ display: "block", overflowWrap: "anywhere" }}
          >
            {user.id}
          </small>
        </p>
        {wallet && (
          <p>
            موجودی فعلی: <b>{money(wallet.balance_toman)}</b> · قابل برداشت:{" "}
            {money(wallet.withdrawable_toman)}
          </p>
        )}
        {error && (
          <Alert severity="error" role="alert">
            {error}
            {submitted.current && (
              <p>
                نتیجه ثبت هنوز قطعی نیست؛ تلاش مجدد با همان اطلاعات انجام
                می‌شود.
              </p>
            )}
          </Alert>
        )}
        {result ? (
          <Alert severity="success" role="status">
            مبلغ <b>{money(result.amount_toman)}</b> با عنوان «{result.title}»
            به کیف پول {result.customer_name} اضافه شد.
            <p>
              {result.withdrawable
                ? "این اعتبار قابل برداشت است."
                : "این اعتبار فقط در سایت قابل خرج است."}
            </p>
            <p>موجودی پس از ثبت: {money(result.balance_toman)}</p>
            <small dir="ltr" style={{ overflowWrap: "anywhere" }}>
              {result.id}
            </small>
          </Alert>
        ) : (
          <form
            id="wallet-credit-form"
            onSubmit={(e) => {
              e.preventDefault();
              void apply();
            }}
          >
            <fieldset
              disabled={busy || !!submitted.current || !wallet}
              style={{
                border: 0,
                padding: 0,
                minWidth: 0,
                display: "grid",
                gap: 16,
                marginTop: 16,
              }}
            >
              <TextField
                label="مبلغ شارژ (تومان)"
                value={form.amount_toman}
                required
                inputProps={{ inputMode: "decimal", maxLength: 18 }}
                onChange={(e) =>
                  setForm((f) => ({ ...f, amount_toman: e.target.value }))
                }
              />
              <TextField
                label="عنوان شارژ"
                value={form.title}
                required
                inputProps={{ maxLength: 120 }}
                onChange={(e) =>
                  setForm((f) => ({ ...f, title: e.target.value }))
                }
              />
              <TextField
                label="توضیحات و دلیل شارژ"
                value={form.description}
                required
                multiline
                minRows={3}
                inputProps={{ maxLength: 1000 }}
                onChange={(e) =>
                  setForm((f) => ({ ...f, description: e.target.value }))
                }
              />
              <FormControlLabel
                label="این شارژ قابل برداشت باشد"
                control={
                  <Checkbox
                    checked={form.withdrawable}
                    onChange={(e) =>
                      setForm((f) => ({ ...f, withdrawable: e.target.checked }))
                    }
                  />
                }
              />
              <Alert severity={form.withdrawable ? "info" : "warning"}>
                {form.withdrawable
                  ? "وجه قابل خرید و برداشت خواهد بود."
                  : "فقط قابل خرید در سایت؛ لغو سفارش هم آن را قابل برداشت نمی‌کند."}
              </Alert>
            </fieldset>
            {!!wallet?.credits.length && (
              <details style={{ marginTop: 16 }}>
                <summary>اعتبارهای قبلی این مشتری</summary>
                {[...wallet.credits]
                  .reverse()
                  .slice(0, 10)
                  .map((c) => (
                    <p key={c.id}>
                      <b>{c.title}</b> · {money(c.amount_toman)} ·{" "}
                      {c.withdrawable ? "قابل برداشت" : "فقط‌خرید"} ·{" "}
                      {new Date(c.at).toLocaleString("fa-IR")}
                    </p>
                  ))}
              </details>
            )}
          </form>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={close} disabled={busy}>
          بستن
        </Button>
        {!result && (
          <Button
            variant="contained"
            type="submit"
            form="wallet-credit-form"
            disabled={busy || !wallet}
          >
            {submitted.current ? "تلاش مجدد همان شارژ" : "ثبت شارژ کیف پول"}
          </Button>
        )}
      </DialogActions>
    </Dialog>
  );
}
