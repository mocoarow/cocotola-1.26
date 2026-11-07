import { useRef } from "react";
import { useTranslation } from "react-i18next";
import { useFetcher } from "react-router";
import { Button } from "~/components/ui/button";
import { SheetFooter } from "~/components/ui/sheet";
import {
  DEFAULT_SOURCE_LANG,
  DEFAULT_TARGET_LANG,
  OrderingFields,
  type OrderingFieldsDefaults,
  useOrderingNotation,
} from "./ordering-fields";

const ADD_DEFAULTS: OrderingFieldsDefaults = {
  sourceText: "",
  sourceLang: DEFAULT_SOURCE_LANG,
  targetLang: DEFAULT_TARGET_LANG,
  explanation: "",
  tags: "",
};

export function OrderingAddForm({ onAdded }: { onAdded: () => void }) {
  const fetcher = useFetcher<{ ok: boolean; added?: boolean }>();
  const { t } = useTranslation();
  const notation = useOrderingNotation("");
  const isSubmitting = fetcher.state !== "idle";

  const added = useRef(false);
  if (fetcher.state === "idle" && fetcher.data?.added && !added.current) {
    added.current = true;
    onAdded();
  }
  if (fetcher.state === "submitting") {
    added.current = false;
  }

  return (
    <fetcher.Form method="post" className="flex flex-1 flex-col gap-4 overflow-y-auto px-4">
      <input type="hidden" name="intent" value="addOrdering" />
      <OrderingFields idPrefix="ordering-" defaults={ADD_DEFAULTS} state={notation} />
      <SheetFooter>
        <Button type="submit" disabled={isSubmitting || !notation.result.ok}>
          {isSubmitting ? t("workbooks.addQuestion.adding") : t("workbooks.addQuestion.submit")}
        </Button>
      </SheetFooter>
    </fetcher.Form>
  );
}
