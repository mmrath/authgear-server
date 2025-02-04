import React from "react";
import { gql } from "@apollo/client";

import { client } from "../../portal/apollo";
import {
  PortalAPIApp,
  PortalAPIAppConfig,
  PortalAPISecretConfig,
} from "../../../types";
import {
  UpdateAppAndSecretConfigMutation,
  UpdateAppAndSecretConfigMutationVariables,
} from "./__generated__/UpdateAppAndSecretConfigMutation";
import { useGraphqlMutation } from "../../../hook/graphql";

// This mutation must fetch the the same set (or super set) fields of AppAndSecretConfigQuery.
// Otherwise, after the mutation, that query will be refetched by Apollo.
const updateAppAndSecretConfigMutation = gql`
  mutation UpdateAppAndSecretConfigMutation(
    $appID: ID!
    $appConfig: AppConfig!
    $secretConfig: SecretConfigInput
  ) {
    updateApp(
      input: {
        appID: $appID
        appConfig: $appConfig
        secretConfig: $secretConfig
      }
    ) {
      app {
        id
        effectiveAppConfig
        rawAppConfig
        secretConfig {
          oauthClientSecrets {
            alias
            clientSecret
          }
          webhookSecret {
            secret
          }
          adminAPISecrets {
            keyID
            createdAt
            publicKeyPEM
            privateKeyPEM
          }
          smtpSecret {
            host
            port
            username
            password
          }
        }
      }
    }
  }
`;

// sanitizeSecretConfig makes sure the return value does not contain fields like __typename.
// The GraphQL runtime will complain about unknown fields.
function sanitizeSecretConfig(
  secretConfig: PortalAPISecretConfig
): PortalAPISecretConfig {
  return {
    oauthClientSecrets:
      secretConfig.oauthClientSecrets?.map((oauthClientSecret) => {
        return {
          alias: oauthClientSecret.alias,
          clientSecret: oauthClientSecret.clientSecret,
        };
      }) ?? null,
    smtpSecret:
      secretConfig.smtpSecret != null
        ? {
            host: secretConfig.smtpSecret.host,
            port: secretConfig.smtpSecret.port,
            username: secretConfig.smtpSecret.username,
            password: secretConfig.smtpSecret.password,
          }
        : null,
  };
}

export function useUpdateAppAndSecretConfigMutation(appID: string): {
  updateAppAndSecretConfig: (
    appConfig: PortalAPIAppConfig,
    secretConfig: PortalAPISecretConfig
  ) => Promise<PortalAPIApp | null>;
  loading: boolean;
  error: unknown;
  resetError: () => void;
} {
  const [mutationFunction, { error, loading }, resetError] = useGraphqlMutation<
    UpdateAppAndSecretConfigMutation,
    UpdateAppAndSecretConfigMutationVariables
  >(updateAppAndSecretConfigMutation, { client });
  const updateAppAndSecretConfig = React.useCallback(
    async (
      appConfig: PortalAPIAppConfig,
      secretConfig: PortalAPISecretConfig
    ) => {
      const result = await mutationFunction({
        variables: {
          appID,
          appConfig: appConfig,
          secretConfig: sanitizeSecretConfig(secretConfig),
        },
      });
      return result.data?.updateApp.app ?? null;
    },
    [appID, mutationFunction]
  );
  return { updateAppAndSecretConfig, error, loading, resetError };
}
