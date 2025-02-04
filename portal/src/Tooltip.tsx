import React, { useRef, useState, useCallback } from "react";
import cn from "classnames";
import {
  TooltipHost,
  ITooltipProps,
  Icon,
  DirectionalHint,
} from "@fluentui/react";

import styles from "./Tooltip.module.scss";
import { FormattedMessage } from "@oursky/react-messageformat";

interface TooltipProps {
  className?: string;
  tooltipMessageId: string;
  children?: React.ReactNode;
}

export interface UseTooltipTargetElementResult {
  id: string;
  setRef: React.RefCallback<unknown>;
  targetElement: HTMLElement | undefined;
}

export function useTooltipTargetElement(): UseTooltipTargetElementResult {
  const { current: id } = useRef(String(Math.random()));
  const [targetElement, setTargetElement] = useState<HTMLElement | null>(null);
  const setRef = useCallback(
    (ref) => {
      if (ref == null) {
        setTargetElement(null);
      } else {
        setTargetElement(document.getElementById(id));
      }
    },
    [id, setTargetElement]
  );
  return {
    id,
    setRef,
    targetElement: targetElement ?? undefined,
  };
}

const Tooltip: React.FC<TooltipProps> = function Tooltip(props: TooltipProps) {
  const { className, tooltipMessageId, children } = props;
  const tooltipProps: ITooltipProps = React.useMemo(() => {
    return {
      // eslint-disable-next-line react/no-unstable-nested-components
      onRenderContent: () => (
        <div className={styles.tooltip}>
          <span>
            <FormattedMessage id={tooltipMessageId} />
          </span>
        </div>
      ),
    };
  }, [tooltipMessageId]);

  return (
    <div className={cn(className, styles.root)}>
      <TooltipHost
        tooltipProps={tooltipProps}
        directionalHint={DirectionalHint.bottomCenter}
      >
        {children ? (
          children
        ) : (
          <Icon className={styles.infoIcon} iconName={"info"} />
        )}
      </TooltipHost>
    </div>
  );
};

export default Tooltip;
