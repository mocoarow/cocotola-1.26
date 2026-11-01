import { useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "~/components/ui/button";

export function AudioPlayButton({ url, label }: { url: string; label: string }) {
  const { t } = useTranslation();
  const audioRef = useRef<HTMLAudioElement | null>(null);
  const [failed, setFailed] = useState(false);

  function handlePlay() {
    const el = audioRef.current;
    if (!el) return;
    // Restart from the beginning so repeated taps always play from 0:00.
    el.currentTime = 0;
    setFailed(false);
    el.play().catch((reason) => {
      // Common causes: autoplay rejection by the browser, network error, decode error.
      console.warn("audio playback failed", reason);
      setFailed(true);
    });
  }

  return (
    <div className="inline-flex flex-col items-end gap-1">
      <Button type="button" variant="outline" size="sm" aria-label={label} onClick={handlePlay}>
        <span aria-hidden="true">▶</span>
        <span className="ml-1 text-xs">{label}</span>
        <audio ref={audioRef} src={url} preload="none" className="hidden">
          <track kind="captions" />
        </audio>
      </Button>
      {failed && (
        <span role="alert" className="text-xs text-red-600">
          {t("workbooks.study.audioPlayFailed")}
        </span>
      )}
    </div>
  );
}
