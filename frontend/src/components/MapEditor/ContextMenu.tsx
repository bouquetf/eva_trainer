import React from "react";

export interface ContextMenuItem {
  label: string;
  action?: () => void;
  disabled?: boolean;
  children?: ContextMenuItem[];
}

interface ContextMenuProps {
  x: number;
  y: number;
  items: ContextMenuItem[];
  onClose: () => void;
}

export const ContextMenu: React.FC<ContextMenuProps> = ({ x, y, items, onClose }) => {
  const [hoveredItem, setHoveredItem] = React.useState<number | null>(null);
  const closeTimerRef = React.useRef<ReturnType<typeof setTimeout> | null>(null);

  React.useEffect(() => {
    const handleClick = () => onClose();
    window.addEventListener("click", handleClick);
    return () => window.removeEventListener("click", handleClick);
  }, [onClose]);

  const handleSubmenuEnter = React.useCallback((index: number) => {
    if (closeTimerRef.current) {
      clearTimeout(closeTimerRef.current);
      closeTimerRef.current = null;
    }
    setHoveredItem(index);
  }, []);

  const handleMenuLeave = React.useCallback(() => {
    closeTimerRef.current = setTimeout(() => {
      setHoveredItem(null);
    }, 300);
  }, []);

  const handleItemClick = React.useCallback((item: ContextMenuItem) => {
    if (item.action && !item.disabled && !item.children) {
      item.action();
      onClose();
    }
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
      onMouseLeave={handleMenuLeave}
    >
      {items.map((item, index) => (
        <div
          key={index}
          onClick={() => handleItemClick(item)}
          style={{
            padding: "8px 15px",
            cursor: item.disabled ? "default" : "pointer",
            color: item.disabled ? "#666" : "#e0e0e0",
            fontSize: "14px",
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            background: hoveredItem === index ? "#333" : "transparent",
            position: "relative",
          }}
          onMouseEnter={() => {
            if (!item.disabled) {
              handleSubmenuEnter(index);
            }
          }}
        >
          <span>{item.label}</span>
          {item.children && <span style={{ marginLeft: "20px", color: "#888" }}>&gt;</span>}
          {item.children && hoveredItem === index && (
            <div
              style={{
                position: "absolute",
                top: 0,
                left: "100%",
                background: "#1e1e1e",
                border: "1px solid #333",
                borderRadius: "4px",
                padding: "5px 0",
                zIndex: 1001,
                boxShadow: "0 2px 10px rgba(0,0,0,0.5)",
                minWidth: "120px",
              }}
              onMouseEnter={() => handleSubmenuEnter(index)}
              onMouseLeave={handleMenuLeave}
            >
              {item.children.map((child, childIndex) => (
                <div
                  key={childIndex}
                  onClick={(e) => {
                    e.stopPropagation();
                    if (child.action && !child.disabled) {
                      child.action();
                      onClose();
                    }
                  }}
                  style={{
                    padding: "8px 15px",
                    cursor: child.disabled ? "default" : "pointer",
                    color: child.disabled ? "#666" : "#e0e0e0",
                    fontSize: "14px",
                    whiteSpace: "nowrap",
                  }}
                  onMouseEnter={(e) => {
                    if (!child.disabled) (e.currentTarget.style.background = "#333");
                  }}
                  onMouseLeave={(e) => {
                    (e.currentTarget.style.background = "transparent");
                  }}
                >
                  {child.label}
                </div>
              ))}
            </div>
          )}
        </div>
      ))}
    </div>
  );
};
