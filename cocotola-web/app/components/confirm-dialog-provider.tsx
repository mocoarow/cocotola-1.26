import { createContext, useCallback, useContext, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogTitle,
} from "~/components/ui/alert-dialog";
import { Button } from "~/components/ui/button";

interface ConfirmOptions {
  title: string;
  description: string;
  confirmLabel?: string;
  cancelLabel?: string;
}

type ConfirmFn = (options: ConfirmOptions) => Promise<boolean>;

const ConfirmDialogContext = createContext<ConfirmFn | null>(null);

export function useConfirm(): ConfirmFn {
  const confirm = useContext(ConfirmDialogContext);
  if (!confirm) {
    throw new Error("useConfirm must be used within ConfirmDialogProvider");
  }
  return confirm;
}

export function ConfirmDialogProvider({ children }: { children: React.ReactNode }) {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);
  const [options, setOptions] = useState<ConfirmOptions | null>(null);
  const resolveRef = useRef<((value: boolean) => void) | null>(null);
  const openRef = useRef(false);

  const confirm = useCallback<ConfirmFn>((opts) => {
    if (openRef.current) {
      console.warn("useConfirm: another confirm dialog is already open; ignoring this call.");
      return Promise.resolve(false);
    }

    return new Promise<boolean>((resolve) => {
      openRef.current = true;
      resolveRef.current = resolve;
      setOptions(opts);
      setOpen(true);
    });
  }, []);

  function resolve(value: boolean) {
    resolveRef.current?.(value);
    resolveRef.current = null;
    openRef.current = false;
    setOpen(false);
  }

  return (
    <ConfirmDialogContext value={confirm}>
      {children}
      <AlertDialog
        open={open}
        onOpenChange={(nextOpen) => {
          if (!nextOpen) resolve(false);
        }}
      >
        <AlertDialogContent>
          <AlertDialogTitle>{options?.title}</AlertDialogTitle>
          <AlertDialogDescription>{options?.description}</AlertDialogDescription>
          <AlertDialogFooter>
            <Button variant="outline" onClick={() => resolve(false)}>
              {options?.cancelLabel ?? t("common.cancel")}
            </Button>
            <Button variant="destructive" onClick={() => resolve(true)}>
              {options?.confirmLabel ?? t("common.confirm")}
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </ConfirmDialogContext>
  );
}
