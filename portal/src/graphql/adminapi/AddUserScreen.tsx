import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { useNavigate, useParams } from "react-router-dom";
import deepEqual from "deep-equal";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { ChoiceGroup, IChoiceGroupOption, Label, Text } from "@fluentui/react";
import { useAppConfigQuery } from "../portal/query/appConfigQuery";
import { useCreateUserMutation } from "./mutations/createUserMutation";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import PasswordField from "../../PasswordField";
import { useTextField } from "../../hook/useInput";
import {
  LoginIDKeyType,
  loginIDKeyTypes,
  PortalAPIAppConfig,
} from "../../types";
import { ErrorParseRule } from "../../error/parse";
import { SimpleFormModel, useSimpleForm } from "../../hook/useSimpleForm";
import FormTextField from "../../FormTextField";
import FormContainer from "../../FormContainer";

import styles from "./AddUserScreen.module.scss";

interface FormState {
  selectedLoginIDType: LoginIDKeyType | null;
  username: string;
  email: string;
  phone: string;
  password: string;
}

const defaultFormState: FormState = {
  selectedLoginIDType: null,
  username: "",
  email: "",
  phone: "",
  password: "",
};

interface AddUserContentProps {
  appConfig: PortalAPIAppConfig | null;
  form: SimpleFormModel<FormState>;
}

const loginIdTypeNameIds: Record<LoginIDKeyType, string> = {
  username: "login-id-key.username",
  email: "login-id-key.email",
  phone: "login-id-key.phone",
};

function determineIsPasswordNeeded(
  appConfig: PortalAPIAppConfig | null,
  loginIdKeySelected: LoginIDKeyType | null
) {
  if (loginIdKeySelected == null) {
    return false;
  }
  const primaryAuthenticators =
    appConfig?.authentication?.primary_authenticators ?? [];
  // password is first one, all login ID types need password
  if (primaryAuthenticators[0] === "password") {
    return true;
  }
  // password is second one, require password if user choose username
  // only id is username -> need password
  if (deepEqual(["oob_otp", "password"], primaryAuthenticators)) {
    return loginIdKeySelected === "username";
  }
  return false;
}

function getLoginIdTypeOptions(appConfig: PortalAPIAppConfig | null) {
  const primaryAuthenticators =
    appConfig?.authentication?.primary_authenticators ?? [];

  // need password authenticator in order to use username to login
  const usernameAllowed = primaryAuthenticators.includes("password");

  const loginIdKeys = appConfig?.identity?.login_id?.keys ?? [];
  const enabledIdentities = appConfig?.authentication?.identities ?? [];
  const isLoginIDIdentityEnabled = enabledIdentities.includes("login_id");

  // if login ID identity is disabled
  // we cannot add login ID identity to new user
  if (!isLoginIDIdentityEnabled) {
    return [];
  }

  const loginIdKeyOptions = new Set<LoginIDKeyType>();
  for (const key of loginIdKeys) {
    switch (key.type) {
      case "username": {
        if (usernameAllowed) {
          loginIdKeyOptions.add("username");
        }
        break;
      }
      case "email":
        loginIdKeyOptions.add("email");
        break;
      case "phone":
        loginIdKeyOptions.add("phone");
        break;
      default:
        break;
    }
  }
  return Array.from(loginIdKeyOptions);
}

const AddUserContent: React.FC<AddUserContentProps> = function AddUserContent(
  props: AddUserContentProps
) {
  const {
    appConfig,
    form: { state, setState },
  } = props;
  const { renderToString } = useContext(Context);

  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      { to: "../..", label: <FormattedMessage id="UsersScreen.title" /> },
      { to: ".", label: <FormattedMessage id="AddUserScreen.title" /> },
    ];
  }, []);

  const { username, email, phone, password, selectedLoginIDType } = state;

  const { onChange: onUsernameChange } = useTextField((value) => {
    setState((prev) => ({ ...prev, username: value }));
  });
  const { onChange: onEmailChange } = useTextField((value) => {
    setState((prev) => ({ ...prev, email: value }));
  });
  const { onChange: onPhoneChange } = useTextField((value) => {
    setState((prev) => ({ ...prev, phone: value }));
  });
  const { onChange: onPasswordChange } = useTextField((value) => {
    setState((prev) => ({ ...prev, password: value }));
  });

  const onSelectLoginIdType = useCallback(
    (_event, options?: IChoiceGroupOption) => {
      const loginIdType = (options?.key ?? null) as LoginIDKeyType | null;
      if (!loginIdType || !loginIDKeyTypes.includes(loginIdType)) {
        return;
      }
      setState((prev) => ({ ...prev, selectedLoginIDType: loginIdType }));
    },
    [setState]
  );
  const renderUsernameField = useCallback(() => {
    return (
      <FormTextField
        className={styles.textField}
        value={username}
        onChange={onUsernameChange}
        parentJSONPointer="/"
        fieldName="username"
        errorRules={[
          {
            reason: "ValidationFailed",
            location: "",
            kind: "format",
            errorMessageID: "AddUserScreen.error.invalid-identity",
          },
        ]}
      />
    );
  }, [username, onUsernameChange]);

  const renderEmailField = useCallback(() => {
    return (
      <FormTextField
        className={styles.textField}
        value={email}
        onChange={onEmailChange}
        parentJSONPointer="/"
        fieldName="email"
        errorRules={[
          {
            reason: "ValidationFailed",
            location: "",
            kind: "format",
            errorMessageID: "AddUserScreen.error.invalid-identity",
          },
        ]}
      />
    );
  }, [email, onEmailChange]);

  const renderPhoneField = useCallback(() => {
    return (
      <FormTextField
        className={styles.textField}
        value={phone}
        onChange={onPhoneChange}
        parentJSONPointer="/"
        fieldName="phone"
        errorRules={[
          {
            reason: "ValidationFailed",
            location: "",
            kind: "format",
            errorMessageID: "AddUserScreen.error.invalid-identity",
          },
        ]}
      />
    );
  }, [phone, onPhoneChange]);

  const textFieldRenderer: Record<
    LoginIDKeyType,
    () => React.ReactNode
  > = useMemo(
    () => ({
      username: renderUsernameField,
      email: renderEmailField,
      phone: renderPhoneField,
    }),
    [renderUsernameField, renderEmailField, renderPhoneField]
  );

  const passwordRequired = useMemo(() => {
    return determineIsPasswordNeeded(appConfig, selectedLoginIDType);
  }, [appConfig, selectedLoginIDType]);

  const loginIdTypeOptions: IChoiceGroupOption[] = useMemo(() => {
    const list = getLoginIdTypeOptions(appConfig);
    return list.map((loginIdType) => {
      const messageId = loginIdTypeNameIds[loginIdType];
      const renderTextField =
        selectedLoginIDType === loginIdType
          ? textFieldRenderer[loginIdType]
          : undefined;
      return {
        key: loginIdType,
        text: renderToString(messageId),
        onRenderLabel: (option) => {
          return option ? (
            <div className={styles.identityOption}>
              <Label className={styles.identityOptionLabel}>
                {option.text}
              </Label>
              {renderTextField?.()}
            </div>
          ) : null;
        },
      };
    });
  }, [appConfig, renderToString, textFieldRenderer, selectedLoginIDType]);

  // NOTE: cannot add user identity if none of three field is available
  const canAddUser = loginIdTypeOptions.length > 0;

  const passwordPolicy = useMemo(() => {
    return appConfig?.authenticator?.password?.policy ?? {};
  }, [appConfig]);

  // TODO: improve empty state
  if (!canAddUser) {
    return (
      <Text>
        <FormattedMessage id="AddUserScreen.cannot-add-user" />
      </Text>
    );
  }

  return (
    <div className={styles.root}>
      <NavBreadcrumb items={navBreadcrumbItems} />
      <ChoiceGroup
        className={styles.userInfo}
        styles={{ label: { marginBottom: "15px", fontSize: "14px" } }}
        selectedKey={selectedLoginIDType ?? undefined}
        options={loginIdTypeOptions}
        onChange={onSelectLoginIdType}
        label={renderToString("AddUserScreen.user-info.label")}
      />
      <PasswordField
        textFieldClassName={styles.textField}
        disabled={!passwordRequired}
        label={renderToString("AddUserScreen.password.label")}
        value={password}
        onChange={onPasswordChange}
        passwordPolicy={passwordPolicy}
        parentJSONPointer="/"
        fieldName="password"
      />
    </div>
  );
};

const AddUserScreen: React.FC = function AddUserScreen() {
  const { appID } = useParams();
  const navigate = useNavigate();

  const { effectiveAppConfig, loading, error, refetch } = useAppConfigQuery(
    appID
  );

  const { createUser } = useCreateUserMutation();

  const [
    lastSubmittedType,
    setLastSubmittedType,
  ] = useState<LoginIDKeyType | null>(null);
  const submit = useCallback(
    async (state: FormState) => {
      const loginIDType = state.selectedLoginIDType;
      if (!loginIDType) {
        return;
      }

      const isPasswordRequired = determineIsPasswordNeeded(
        effectiveAppConfig,
        loginIDType
      );
      // FIXME: local password validation error
      /*  const passwordValidationResult =  determineIsPasswordNeeded(effectiveAppConfig, loginIDType)
              ? localValidatePassword(
                  renderToString,
                  effectiveAppConfig?.authenticator?.password?.policy ?? {},
                  state.password,
              )
              : null;
          if (passwordValidationResult != null) {
            setLocalValidationErrorMessage(passwordValidationResult?.password);
            return;
          }*/
      setLastSubmittedType(loginIDType);
      const identityValue = state[loginIDType];
      const password = isPasswordRequired ? state.password : undefined;

      await createUser({ key: loginIDType, value: identityValue }, password);
    },
    [createUser, effectiveAppConfig]
  );

  const form = useSimpleForm(defaultFormState, submit);

  const errorRules: ErrorParseRule[] = useMemo(
    () => [
      {
        reason: "InvariantViolated",
        kind: "DuplicatedIdentity",
        errorMessageID: "AddUserScreen.error.duplicated-identity",
        field: lastSubmittedType ?? "",
      },
    ],
    [lastSubmittedType]
  );

  const canSave = form.state.selectedLoginIDType != null;
  const saveButtonProps = useMemo(
    () => ({
      labelId: "AddUserScreen.add-user.label",
      iconName: "Add",
    }),
    []
  );

  useEffect(() => {
    if (form.isSubmitted) {
      navigate("..");
    }
  }, [form.isSubmitted, navigate]);

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <FormContainer
      form={form}
      canSave={canSave}
      saveButtonProps={saveButtonProps}
      errorRules={errorRules}
    >
      <AddUserContent form={form} appConfig={effectiveAppConfig} />
    </FormContainer>
  );
};

export default AddUserScreen;
