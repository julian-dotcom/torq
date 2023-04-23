import {
  ArrowSyncFilled as ProcessingIcon,
  CheckmarkRegular as SuccessIcon,
  ErrorCircleRegular as FailedIcon,
  Note20Regular as NoteIcon,
  Molecule20Regular as PeersIcon,
} from "@fluentui/react-icons";
import { useGetNodeConfigurationsQuery } from "apiSlice";
import { useEffect, useState } from "react";
import Button, { ButtonWrapper, ColorVariant } from "components/buttons/Button";
import ProgressHeader, { ProgressStepState, Step } from "features/progressTabs/ProgressHeader";
import ProgressTabs, { ProgressTabContainer } from "features/progressTabs/ProgressTab";
import styles from "./peers.module.scss";
import { useNavigate } from "react-router";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import useTranslations from "services/i18n/useTranslations";
import { nodeConfiguration } from "apiTypes";
import { SelectOptions } from "features/forms/Select";
import { ActionMeta } from "react-select";
import classNames from "classnames";
import FormRow from "features/forms/FormWrappers";
import Note, { NoteType } from "features/note/Note";
import { Select, TextArea } from "components/forms/forms";
import { InputSizeVariant } from "components/forms/input/variants";
import { userEvents } from "utils/userEvents";
import { FormErrors, mergeServerError, ServerErrorType } from "components/errors/errors";
import ErrorSummary from "components/errors/ErrorSummary";
import { useConnectPeerMutation } from "./peersApi";
import { useAppSelector } from "store/hooks";
import { selectActiveNetwork } from "../network/networkSlice";
import clone from "clone";

const updateStatusClass = {
  PROCESSING: styles.processing,
  FAILED: styles.failed,
  SUCCEEDED: styles.success,
};

const updateStatusIcon = {
  PROCESSING: <ProcessingIcon />,
  FAILED: <FailedIcon />,
  SUCCEEDED: <SuccessIcon />,
  NOTE: <NoteIcon />,
};
function isOption(result: unknown): result is SelectOptions & { value: number } {
  return (
    result !== null &&
    typeof result === "object" &&
    "value" in result &&
    "label" in result &&
    typeof (result as SelectOptions).value === "number"
  );
}

function ConnectPeerModal() {
  const { t } = useTranslations();
  const { track } = userEvents();
  const navigate = useNavigate();

  const [nodeConfigurationOptions, setNodeConfigurationOptions] = useState<Array<SelectOptions>>();
  const [connectState, setConnectState] = useState(ProgressStepState.active);
  const [stepIndex, setStepIndex] = useState(0);

  const [selectedNodeId, setSelectedNodeId] = useState<number | undefined>();
  const [nodePubKey, setNodePubKey] = useState<string>("");
  const [connectionString, setConnectionString] = useState<string>("");
  const [host, setHost] = useState<string>("");
  const activeNetwork = useAppSelector(selectActiveNetwork);
  const [resultState, setResultState] = useState(ProgressStepState.disabled);
  const [formErrorState, setFormErrorState] = useState({} as FormErrors);

  const [connectPeer, response] = useConnectPeerMutation();

  const { data: nodeConfigurations } = useGetNodeConfigurationsQuery();

  useEffect(() => {
    if (nodeConfigurations) {
      const options = nodeConfigurations.map((node: nodeConfiguration) => {
        return { label: node.name, value: node.nodeId };
      });
      setNodeConfigurationOptions(options);
      setSelectedNodeId(options[0].value);
    }
  }, [nodeConfigurations]);

  useEffect(() => {
    if (response && response.isError && response.error && "data" in response.error && response.error.data) {
      const mergedErrors = mergeServerError(response.error.data as ServerErrorType, clone(formErrorState));
      setFormErrorState(mergedErrors);
      setResultState(ProgressStepState.error);
    }
    if (response && response.isLoading) {
      setResultState(ProgressStepState.processing);
    }
    if (response.isSuccess) {
      setResultState(ProgressStepState.completed);
    }
  }, [response]);

  function handleNodeSelection(value: number) {
    setSelectedNodeId(value);
  }

  const closeAndReset = () => {
    setStepIndex(0);
    setConnectState(ProgressStepState.active);

    setNodePubKey("");
    setConnectionString("");
    setResultState(ProgressStepState.disabled);
  };

  function handleConnectPeer() {
    if (!selectedNodeId && !connectionString) return;

    setStepIndex(1);
    setResultState(ProgressStepState.processing);
    track("Connect Peer", {
      torqNodeId: selectedNodeId,
    });
    connectPeer({
      nodeId: selectedNodeId ?? 0,
      network: activeNetwork,
      connectionString: connectionString,
    });
  }

  return (
    <PopoutPageTemplate title={t.peersPage.connectPeer} show={true} onClose={() => navigate(-1)} icon={<PeersIcon />}>
      <ProgressHeader modalCloseHandler={closeAndReset}>
        <Step label={"Connect"} state={connectState} last={false} />
        <Step label={"Result"} state={resultState} last={true} />
      </ProgressHeader>

      <ProgressTabs showTabIndex={stepIndex}>
        <ProgressTabContainer>
          <FormRow>
            <Select
              intercomTarget={"connect-peer-node-select"}
              label={t.yourNode}
              onChange={(newValue: unknown, _: ActionMeta<unknown>) => {
                // Check if newValue is of type SelectOptions
                if (isOption(newValue)) {
                  const selectOptions = newValue as SelectOptions;
                  handleNodeSelection(selectOptions?.value as number);
                }
              }}
              options={nodeConfigurationOptions}
              value={nodeConfigurationOptions?.find((option) => option.value === selectedNodeId)}
            />
          </FormRow>
          <FormRow>
            <div className={styles.openChannelTableSingle}>
              <div className={styles.input}>
                <TextArea
                  label={t.ConnectionString}
                  helpText={t.NodeConnectionStringHelp}
                  sizeVariant={InputSizeVariant.normal}
                  value={connectionString}
                  rows={4}
                  placeholder={"03aab7e9327716ee946b8fbfae039b01235356549e72c5cca113ea67893d0821e5@123.1.3.65:9735"}
                  onChange={(e) => {
                    setConnectionString(e.target.value);
                    if (!e.target.value) {
                      setNodePubKey("");
                    }
                    if (e.target.value) {
                      const split = e.target.value.split("@");
                      split[0] && setNodePubKey(split[0]);
                      split[1] && setHost(split[1]);
                    }
                  }}
                />
              </div>
            </div>
          </FormRow>
          <ButtonWrapper
            rightChildren={
              <Button
                intercomTarget={"connect-peer-confirm-button"}
                disabled={host == "" || nodePubKey == "" || selectedNodeId === undefined}
                onClick={() => {
                  setStepIndex(1);
                  setConnectState(ProgressStepState.completed);
                  handleConnectPeer();
                }}
                buttonColor={ColorVariant.primary}
              >
                {t.confirm}
              </Button>
            }
          />
        </ProgressTabContainer>
        <ProgressTabContainer>
          <div
            className={classNames(
              styles.peerResultIconWrapper,
              { [styles.failed]: !response.data },
              updateStatusClass[response.isLoading ? "PROCESSING" : response.isError ? "FAILED" : "SUCCEEDED"]
            )}
          >
            {updateStatusIcon[response.isLoading ? "PROCESSING" : response.isSuccess ? "SUCCEEDED" : "FAILED"]}
          </div>
          {response.isLoading && (
            <Note title={t.Processing} icon={<ProcessingIcon />} noteType={NoteType.warning}>
              {t.openCloseChannel.processingClose}
            </Note>
          )}
          <div className={styles.peersResultDetails}>
            {response.isSuccess && (
              <Note title={t.note} icon={<NoteIcon />} noteType={NoteType.info}>
                {t.peersPage.confirmationConnectPeer}
              </Note>
            )}
            <ErrorSummary errors={formErrorState} />
            <ButtonWrapper
              className={styles.resetButton}
              rightChildren={
                <Button
                  intercomTarget={"connect-peer-another-peer-button"}
                  onClick={() => {
                    closeAndReset();
                  }}
                  buttonColor={ColorVariant.primary}
                >
                  {t.peersPage.connectToAnotherPeer}
                </Button>
              }
            />
          </div>
        </ProgressTabContainer>
      </ProgressTabs>
    </PopoutPageTemplate>
  );
}

export default ConnectPeerModal;
