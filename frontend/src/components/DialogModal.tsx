import type {PropsWithChildren} from "react";

interface DialogModalProps extends PropsWithChildren {
    onClose: () => void;
    open: boolean;
    title: string;
    width?: string;
}

export function DialogModal({children, onClose, open, title, width = "960px"}: DialogModalProps) {
    if (!open) {
        return null;
    }

    return (
        <div aria-hidden="true" className="dialog-backdrop" onClick={onClose}>
            <section
                aria-labelledby="dialog-title"
                aria-modal="true"
                className="dialog-shell"
                onClick={(event) => event.stopPropagation()}
                role="dialog"
                style={{maxWidth: width}}
            >
                <header className="dialog-header">
                    <div>
                        <p className="section-kicker">工作台弹窗</p>
                        <h2 id="dialog-title" className="surface-title">{title}</h2>
                    </div>
                    <button className="ghost-button" onClick={onClose} type="button">
                        关闭
                    </button>
                </header>
                <div className="dialog-content">{children}</div>
            </section>
        </div>
    );
}
