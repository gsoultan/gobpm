/**
 * Configuring a sub-process whose steps are driven by a person rather than by
 * the diagram.
 *
 * An ordinary sub-process runs from its own start event and follows its own
 * arrows. An ad-hoc one has neither: it holds the process where it is while a
 * knowledge worker runs the steps inside it, in whatever order the work turns
 * out to need, as many times as it needs, until a completion condition says the
 * work is finished. That is the BPMN element for investigation, triage and case
 * work — anything where the order is not known when the diagram is drawn.
 *
 * The engine has executed these for a while. Nothing in the designer could
 * produce one, so the only way to get one was to import a file that already had
 * one. These are the rules that decide whether what somebody has configured
 * will actually work, kept here rather than in the panel so they can be tested
 * without rendering anything.
 */

export interface AdHocConfig {
  isAdHoc?: boolean;
  completionCondition?: string;
  isEventSubProcess?: boolean;
}

/** A problem worth telling the author about before they deploy. */
export interface AdHocWarning {
  /** Which field it is about, so the panel can put it next to that field. */
  field: 'steps' | 'completionCondition' | 'isEventSubProcess';
  severity: 'error' | 'warning';
  message: string;
}

/**
 * The field changes that turning ad-hoc mode on or off implies.
 *
 * Turning it off clears the completion condition, because that field means
 * something different on an ordinary sub-process — nothing — and a value left
 * behind in it would be invisible in the panel and still be deployed.
 */
export function adHocToggle(enabled: boolean): AdHocConfig {
  return enabled ? { isAdHoc: true } : { isAdHoc: false, completionCondition: '' };
}

/**
 * What is wrong with this configuration, if anything.
 *
 * `stepCount` is how many steps are drawn inside the sub-process. It is a
 * parameter rather than something read here because the panel knows the diagram
 * and this module deliberately does not.
 */
export function adHocWarnings(config: AdHocConfig, stepCount: number): AdHocWarning[] {
  if (!config.isAdHoc) {
    return [];
  }

  const warnings: AdHocWarning[] = [];

  if (stepCount === 0) {
    warnings.push({
      field: 'steps',
      severity: 'error',
      message:
        'There are no steps inside this sub-process, so there is nothing for anyone to run. ' +
        'Drag the steps this work is made of into it.',
    });
  }

  if (!config.completionCondition?.trim()) {
    warnings.push({
      field: 'completionCondition',
      severity: 'warning',
      message:
        'With no condition the process moves on the moment it arrives here, without waiting for ' +
        'any of the steps inside to be run. Say what "finished" means — for example ' +
        'checksDone >= 2.',
    });
  }

  if (config.isEventSubProcess) {
    warnings.push({
      field: 'isEventSubProcess',
      severity: 'error',
      message:
        'An event sub-process is started by an event, and an ad-hoc one is driven by a person. ' +
        'A sub-process cannot be both. Turn one of them off.',
    });
  }

  return warnings;
}

/** Whether anything here would stop the process working, as opposed to being a caution. */
export function adHocBlocks(config: AdHocConfig, stepCount: number): boolean {
  return adHocWarnings(config, stepCount).some((w) => w.severity === 'error');
}
