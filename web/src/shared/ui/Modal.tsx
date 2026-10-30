import type { ReactNode } from "react";
import { Modal as HeroModal } from "@heroui/react";

export function Modal({
  title,
  children,
  onClose,
}: {
  title: string;
  children: ReactNode;
  onClose: () => void;
}) {
  return (
    <HeroModal
      isOpen
      onOpenChange={(open) => {
        if (!open) {
          onClose();
        }
      }}
    >
      <HeroModal.Backdrop className="purchase-modal-backdrop">
        <HeroModal.Container>
          <HeroModal.Dialog className="purchase-modal">
            <HeroModal.CloseTrigger aria-label="Close dialog" />
            <HeroModal.Header className="modal-heading">
              <span className="eyebrow">MAKE ROOM FOR WHAT'S NEXT</span>
              <HeroModal.Heading>{title}</HeroModal.Heading>
            </HeroModal.Header>
            <HeroModal.Body>{children}</HeroModal.Body>
          </HeroModal.Dialog>
        </HeroModal.Container>
      </HeroModal.Backdrop>
    </HeroModal>
  );
}
