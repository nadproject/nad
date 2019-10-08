/* Copyright (C) 2019 Monomax Software Pty Ltd
 *
 * This file is part of Dnote.
 *
 * Dnote is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Dnote is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Dnote.  If not, see <https://www.gnu.org/licenses/>.
 */

import services from 'web/libs/services';
import { DigestRuleData } from 'jslib/operations/types';
import { CreateParams } from 'jslib/services/repetitionRules';
import {
  RECEIVE,
  ADD,
  REMOVE,
  START_FETCHING,
  FINISH_FETCHING,
  RECEIVE_ERROR,
  ReceiveDigestRulesAction,
  ReceiveDigestRulesErrorAction,
  StartFetchingDigestRulesAction,
  FinishFetchingDigestRulesAction,
  AddDigestRuleAction,
  RemoveDigestRuleAction
} from './type';
import { ThunkAction } from '../types';

function receiveDigestRules(
  repetitionRules: DigestRuleData[]
): ReceiveDigestRulesAction {
  return {
    type: RECEIVE,
    data: { repetitionRules }
  };
}

function receiveDigestRulesError(err: string): ReceiveDigestRulesErrorAction {
  return {
    type: RECEIVE_ERROR,
    data: { err }
  };
}

function startFetchingDigestRules(): StartFetchingDigestRulesAction {
  return {
    type: START_FETCHING
  };
}

function finishFetchingDigestRules(): FinishFetchingDigestRulesAction {
  return {
    type: FINISH_FETCHING
  };
}

export const getDigestRules = (): ThunkAction<void> => {
  return dispatch => {
    dispatch(startFetchingDigestRules());

    return services.repetitionRules
      .fetchAll()
      .then(data => {
        dispatch(receiveDigestRules(data));
        dispatch(finishFetchingDigestRules());
      })
      .catch(err => {
        console.log('getDigestRules error', err);
        dispatch(receiveDigestRulesError(err));
      });
  };
};

export function addDigestRule(repetitionRule: DigestRuleData): AddDigestRuleAction {
  return {
    type: ADD,
    data: { repetitionRule }
  };
}

export const createDigestRule = (
  p: CreateParams
): ThunkAction<DigestRuleData> => {
  return dispatch => {
    return services.repetitionRules.create(p).then(data => {
      dispatch(addDigestRule(data));

      return data;
    });
  };
};

export function removeDigestRule(uuid: string): RemoveDigestRuleAction {
  return {
    type: REMOVE,
    data: { uuid }
  };
}
