import React, { useCallback } from "react";
import { ContextMenu } from "./ContextMenu";
import type { EditorState, ContextMenuActionItem } from "../../lib/editor";

interface EditorContextMenuProps {
  editorState: EditorState;
  onAction: (action: string) => void;
  onClose: () => void;
}

export const EditorContextMenu: React.FC<EditorContextMenuProps> = ({
  editorState,
  onAction,
  onClose,
}) => {
  const items = editorState.dialogs.contextMenuActionItems;
  const handleItemClick = useCallback(
    (item: ContextMenuActionItem) => {
      onAction(item.action);
      onClose();
    },
    [onAction, onClose],
  );

  if (!editorState.dialogs.contextMenu) return null;

  return (
    <ContextMenu
      x={editorState.dialogs.contextMenu.x}
      y={editorState.dialogs.contextMenu.y}
      items={items.map((item) => ({
        label: item.label,
        action: item.children
          ? undefined
          : () => handleItemClick(item),
        children: item.children
          ? item.children.map((child) => ({
              label: child.label,
              action: () => handleItemClick(child),
            }))
          : undefined,
      }))}
      onClose={onClose}
    />
  );
};
