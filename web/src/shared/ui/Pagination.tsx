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
        <button
          className="button ghost small"
          disabled={!previous}
          onClick={onPrevious}
        >
          ← Previous
        </button>
        <button
          className="button ghost small"
          disabled={!more}
          onClick={onNext}
        >
          Next →
        </button>
      </div>
    </footer>
  );
}
