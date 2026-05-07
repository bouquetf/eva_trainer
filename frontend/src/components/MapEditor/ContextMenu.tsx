import React from "react";

interface ContextMenuItem {
  label: string;
  action: () => void;
  disabled?: boolean;
}

interface ContextMenuProps {
  x: number;
  y: number;
  items: ContextMenuItem[];
  onClose: () => void;
}

export const ContextMenu: React.FC<ContextMenuProps> = ({ x, y, items, onClose }) => {
  React.useEffect(() => {
    const handleClick = () => onClose();
    window.addEventListener("click", handleClick);
    return () => window.removeEventListener("click", handleClick);
  }, [onClose]);

  return (
    <div
      style={{
        position: "fixed",
        top: y,
        left: x,
        background: "#1e1e1e",
        border: "1px solid #333",
        borderRadius: "4px",
        padding: "5px 0",
        zIndex: 1000,
        boxShadow: "0 2px 10px rgba(0,0,0,0.5)",
        minWidth: "150px",
      }}
      onClick={(e) => e.stopPropagation()}
    >
      {items.map((item, index) => (
        <div
          key={index}
          onClick={() => {
            item.action();
            onClose();
          }}
          style={{
            padding: "8px 15px",
            cursor: item.disabled ? "default" : "pointer",
            color: item.disabled ? "#666" : "#e0e0e0",
            fontSize: "14px",
          }}
          onMouseEnter={(e) => {
            if (!item.disabled) (e.currentTarget.style.background = "#333");
          }}
          onMouseLeave={(e) => {
            (e.currentTarget.style.background = "transparent");
          }}
        >
          {item.label}
        </div>
      ))}
    </div>
  );
};
