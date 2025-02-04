import React, { useCallback, useContext, useMemo } from "react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import {
  ActionButton,
  DetailsListLayoutMode,
  IColumn,
  SelectionMode,
  ShimmeredDetailsList,
} from "@fluentui/react";
import cn from "classnames";

import {
  Collaborator,
  CollaboratorInvitation,
} from "./query/collaboratorsAndInvitationsQuery";

import styles from "./PortalAdminList.module.scss";
import { useSystemConfig } from "../../context/SystemConfigContext";

interface PortalAdminListProps {
  className?: string;
  loading: boolean;
  collaborators: Collaborator[];
  collaboratorInvitations: CollaboratorInvitation[];
  onRemoveCollaboratorClicked: (
    event: React.MouseEvent<unknown>,
    id: string
  ) => void;
  onRemoveCollaboratorInvitationClicked: (
    event: React.MouseEvent<unknown>,
    id: string
  ) => void;
}

interface PortalAdminListCollaboratorItem {
  type: "collaborator";
  id: string;
  createdAt: Date;
  email: string;
}

interface PortalAdminListCollaboratorInvitationItem {
  type: "collaboratorInvitation";
  id: string;
  createdAt: Date;
  email: string;
}

type PortalAdminListItem =
  | PortalAdminListCollaboratorItem
  | PortalAdminListCollaboratorInvitationItem;

function isPortalAdminListCollaboratorItem(
  item: PortalAdminListItem
): item is PortalAdminListCollaboratorItem {
  return item.type === "collaborator";
}

function isPortalAdminListCollaboratorInvitationItem(
  item: PortalAdminListItem
): item is PortalAdminListCollaboratorInvitationItem {
  return item.type === "collaboratorInvitation";
}

const PortalAdminList: React.FC<PortalAdminListProps> =
  function PortalAdminList(props) {
    const {
      className,
      loading,
      collaborators,
      collaboratorInvitations,
      onRemoveCollaboratorClicked,
      onRemoveCollaboratorInvitationClicked,
    } = props;
    const { themes } = useSystemConfig();

    const { renderToString } = useContext(Context);

    const columns: IColumn[] = useMemo(() => {
      return [
        {
          key: "email",
          fieldName: "email",
          name: renderToString("PortalAdminList.column.email"),
          minWidth: 400,
        },
        {
          key: "status",
          fieldName: "status",
          name: renderToString("PortalAdminList.column.status"),
          minWidth: 150,
        },
        {
          key: "action",
          fieldName: "action",
          name: renderToString("PortalAdminList.column.action"),
          minWidth: 150,
        },
      ];
    }, [renderToString]);

    const items: PortalAdminListItem[] = useMemo(() => {
      return [
        ...collaborators.map<PortalAdminListCollaboratorItem>(
          (collaborator) => ({
            type: "collaborator",
            id: collaborator.id,
            createdAt: new Date(collaborator.createdAt),
            email: collaborator.user.email ?? "",
          })
        ),
        ...collaboratorInvitations.map<PortalAdminListCollaboratorInvitationItem>(
          (collaboratorInvitation) => ({
            type: "collaboratorInvitation",
            id: collaboratorInvitation.id,
            createdAt: new Date(collaboratorInvitation.createdAt),
            email: collaboratorInvitation.inviteeEmail,
          })
        ),
      ];
    }, [collaboratorInvitations, collaborators]);

    const onRenderItemColumn = useCallback(
      (item: PortalAdminListItem, _index?: number, column?: IColumn) => {
        switch (column?.key) {
          case "email":
            return <span>{item.email}</span>;
          case "status":
            if (isPortalAdminListCollaboratorItem(item)) {
              return <span className={styles.acceptedStatus}>Accepted</span>;
            }
            return <span className={styles.pendingStatus}>Pending</span>;
          case "action":
            return (
              <ActionButton
                className={styles.actionButton}
                styles={{ flexContainer: { alignItems: "normal" } }}
                theme={themes.destructive}
                onClick={(event) => {
                  if (isPortalAdminListCollaboratorItem(item)) {
                    onRemoveCollaboratorClicked(event, item.id);
                  } else if (
                    isPortalAdminListCollaboratorInvitationItem(item)
                  ) {
                    onRemoveCollaboratorInvitationClicked(event, item.id);
                  }
                }}
              >
                <FormattedMessage id="PortalAdminList.remove" />
              </ActionButton>
            );
          default:
            return null;
        }
      },
      [
        onRemoveCollaboratorClicked,
        onRemoveCollaboratorInvitationClicked,
        themes.destructive,
      ]
    );

    return (
      <div className={cn(styles.root, className)}>
        <ShimmeredDetailsList
          enableShimmer={loading}
          onRenderItemColumn={onRenderItemColumn}
          selectionMode={SelectionMode.none}
          layoutMode={DetailsListLayoutMode.justified}
          columns={columns}
          items={items}
        />
      </div>
    );
  };

export default PortalAdminList;
