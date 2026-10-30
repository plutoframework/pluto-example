import { Button } from "@heroui/react";

export function Pagination({
  count,
  previous,
  more,
  onPrevious,
  onNext,
}: {
  count: number;
  previous: boolean;
  more: boolean;
  onPrevious: () => void;
  onNext: () => void;
}) {
  return (
    <footer className="pagination">
      <span>{count} records on this page</span>
      <div>
        <Button
          variant="outline"
          size="sm"
          isDisabled={!previous}
          onClick={onPrevious}
        >
          ← Previous
        </Button>
        <Button variant="outline" size="sm" isDisabled={!more} onClick={onNext}>
          Next →
        </Button>
      </div>
    </footer>
  );
}
