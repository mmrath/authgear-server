import React, { useCallback } from "react";
import { useParams } from "react-router-dom";
import { FormattedMessage } from "@oursky/react-messageformat";
import produce from "immer";
import FormContainer from "../../FormContainer";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import UserProfileAttributesList, {
  UserProfileAttributesListItem,
} from "../../UserProfileAttributesList";
import {
  PortalAPIAppConfig,
  StandardAttributesAccessControlConfig,
} from "../../types";
import styles from "./StandardAttributesConfigurationScreen.module.scss";

interface FormState {
  standardAttributesItems: StandardAttributesAccessControlConfig[];
}

interface StandardAttributesConfigurationScreenContentProps {
  form: AppConfigFormModel<FormState>;
}

const naturalOrder = [
  "/email",
  "/phone_number",
  "/preferred_username",
  "/name",
  "/given_name",
  "/family_name",
  "/middle_name",
  "/nickname",
  "/profile",
  "/picture",
  "/website",
  "/gender",
  "/birthdate",
  "/zoneinfo",
  "/locale",
  "/address",
];

function constructFormState(config: PortalAPIAppConfig): FormState {
  const items = config.user_profile?.standard_attributes?.access_control ?? [];
  const listedItems = items.filter((a) => naturalOrder.indexOf(a.pointer) >= 0);
  listedItems.sort((a, b) => {
    const ia = naturalOrder.indexOf(a.pointer);
    const ib = naturalOrder.indexOf(b.pointer);
    return ia - ib;
  });
  return {
    standardAttributesItems: listedItems,
  };
}

function constructConfig(
  rawConfig: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState,
  effectiveConfig: PortalAPIAppConfig
): PortalAPIAppConfig {
  const modifiedEffectiveConfig = produce(
    effectiveConfig,
    (effectiveConfig) => {
      effectiveConfig.user_profile ??= {};
      effectiveConfig.user_profile.standard_attributes ??= {};
      for (const accessControl of effectiveConfig.user_profile
        .standard_attributes.access_control ?? []) {
        for (const item of currentState.standardAttributesItems) {
          if (accessControl.pointer === item.pointer) {
            accessControl.access_control = item.access_control;
          }
        }
      }
    }
  );

  const accessControl =
    modifiedEffectiveConfig.user_profile?.standard_attributes?.access_control;
  return produce(rawConfig, (rawConfig) => {
    rawConfig.user_profile ??= {};
    rawConfig.user_profile.standard_attributes ??= {};
    rawConfig.user_profile.standard_attributes.access_control = accessControl;
  });
}

const StandardAttributesConfigurationScreenContent: React.FC<StandardAttributesConfigurationScreenContentProps> =
  function StandardAttributesConfigurationScreenContent(props) {
    const { state, setState } = props.form;
    const onChangeItems = useCallback(
      (newItems: UserProfileAttributesListItem[]) => {
        setState((prev) => {
          return {
            ...prev,
            standardAttributesItems: newItems,
          };
        });
      },
      [setState]
    );
    return (
      <>
        <ScreenContent>
          <ScreenTitle className={styles.widget}>
            <FormattedMessage id="StandardAttributesConfigurationScreen.title" />
          </ScreenTitle>
          <div className={styles.widget}>
            <UserProfileAttributesList
              items={state.standardAttributesItems}
              onChangeItems={onChangeItems}
            />
          </div>
        </ScreenContent>
      </>
    );
  };

const StandardAttributesConfigurationScreen: React.FC =
  function StandardAttributesConfigurationScreen() {
    const { appID } = useParams();
    const form = useAppConfigForm(appID, constructFormState, constructConfig);

    if (form.isLoading) {
      return <ShowLoading />;
    }

    if (form.loadError) {
      return <ShowError error={form.loadError} onRetry={form.reload} />;
    }

    return (
      <FormContainer form={form}>
        <StandardAttributesConfigurationScreenContent form={form} />
      </FormContainer>
    );
  };

export default StandardAttributesConfigurationScreen;
