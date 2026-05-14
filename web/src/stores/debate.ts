import { ref } from "vue";
import { defineStore } from "pinia";
import type { Debate, Argument, DebateVote, VoteHistory } from "@/types";
import { useAuthStore } from "@/stores/auth";

const API_URL = import.meta.env.VITE_API_URL || "/api";

export const useDebateStore = defineStore("debate", () => {
  const debates = ref<Debate[]>([]);
  const debatesTotal = ref(0);
  const currentDebate = ref<Debate | null>(null);
  const argumentList = ref<Argument[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);
  // Track current user's votes on arguments
  const userVotes = ref<Record<string, number>>({});

  const authHeaders = () => {
    const authStore = useAuthStore();
    return { Authorization: `Bearer ${authStore.token}` };
  };

  const fetchDebates = async (
    params: {
      status?: "open" | "concluded" | "archived";
      tag?: string;
      page?: number;
      limit?: number;
    } = {},
  ) => {
    loading.value = true;
    error.value = null;
    try {
      const query = new URLSearchParams();
      if (params.status) query.set("status", params.status);
      if (params.tag) query.set("tag", params.tag);
      if (params.page) query.set("page", String(params.page));
      if (params.limit) query.set("limit", String(params.limit));

      const res = await fetch(`${API_URL}/debate/topics?${query}`);
      if (res.ok) {
        const data = await res.json();
        debates.value = data.data || [];
        debatesTotal.value = data.total || 0;
      }
    } catch (e) {
      error.value = "Failed to fetch debates";
      console.error(e);
    } finally {
      loading.value = false;
    }
  };

  const fetchDebate = async (id: string) => {
    loading.value = true;
    error.value = null;
    try {
      const res = await fetch(`${API_URL}/debate/topics/${id}`);
      if (res.ok) {
        const data = await res.json();
        currentDebate.value = data.data;
        return data.data as Debate;
      } else {
        error.value = "Debate not found";
        return null;
      }
    } catch (e) {
      error.value = "Failed to fetch debate";
      console.error(e);
      return null;
    } finally {
      loading.value = false;
    }
  };

  const fetchArguments = async (debateId: string) => {
    loading.value = true;
    error.value = null;
    try {
      const authStore = useAuthStore();
      const res = await fetch(`${API_URL}/debate/topics/${debateId}/arguments`, {
        headers: authStore.isAuthenticated ? authHeaders() : {},
      });
      if (res.ok) {
        const data = await res.json();
        argumentList.value = data.data || [];
        userVotes.value = data.user_votes || {};
        return data.data as Argument[];
      } else {
        error.value = "Failed to fetch arguments";
        userVotes.value = {};
        return [];
      }
    } catch (e) {
      error.value = "Failed to fetch arguments";
      console.error(e);
      userVotes.value = {};
      return [];
    } finally {
      loading.value = false;
    }
  };

  const createDebate = async (payload: {
    title: string;
    description: string;
    content: string;
    tags: string[];
  }): Promise<Debate | null> => {
    try {
      const res = await fetch(`${API_URL}/debate/topics`, {
        method: "POST",
        headers: { "Content-Type": "application/json", ...authHeaders() },
        body: JSON.stringify(payload),
      });
      if (res.ok) {
        const data = await res.json();
        return data.data as Debate;
      } else {
        const error = await res.json();
        throw new Error(error.error || "Failed to create debate");
      }
    } catch (e) {
      console.error("Failed to create debate", e);
      return null;
    }
  };

  const updateDebate = async (
    id: string,
    payload: {
      title: string;
      description: string;
      content: string;
      tags: string[];
    },
  ): Promise<Debate | null> => {
    try {
      const res = await fetch(`${API_URL}/debate/topics/${id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json", ...authHeaders() },
        body: JSON.stringify(payload),
      });
      if (res.ok) {
        const data = await res.json();
        return data.data as Debate;
      } else {
        const error = await res.json();
        throw new Error(error.error || "Failed to update debate");
      }
    } catch (e) {
      console.error("Failed to update debate", e);
      return null;
    }
  };

  const deleteDebate = async (id: string): Promise<boolean> => {
    try {
      const res = await fetch(`${API_URL}/debate/topics/${id}`, {
        method: "DELETE",
        headers: authHeaders(),
      });
      return res.ok;
    } catch (e) {
      console.error("Failed to delete debate", e);
      return false;
    }
  };

  const concludeDebate = async (
    id: string,
    payload: { conclusion_type: 'yes' | 'no' | 'inconclusive'; conclusion_summary?: string },
  ): Promise<Debate | null> => {
    try {
      const res = await fetch(`${API_URL}/debate/topics/${id}/conclude`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', ...authHeaders() },
        body: JSON.stringify(payload),
      })
      if (res.ok) {
        const data = await res.json()
        return data.data as Debate
      } else {
        const error = await res.json()
        throw new Error(error.error || 'Failed to conclude debate')
      }
    } catch (e) {
      console.error('Failed to conclude debate', e)
      return null
    }
  }

  const reopenDebate = async (id: string): Promise<Debate | null> => {
    try {
      const res = await fetch(`${API_URL}/debate/topics/${id}/reopen`, {
        method: 'POST',
        headers: authHeaders(),
      })
      if (res.ok) {
        const data = await res.json()
        return data.data as Debate
      } else {
        const error = await res.json()
        throw new Error(error.error || 'Failed to reopen debate')
      }
    } catch (e) {
      console.error('Failed to reopen debate', e)
      return null
    }
  }

  const voteToConclude = async (
    id: string,
  ): Promise<{ conclude_vote_count: number; conclude_threshold: number; auto_concluded: boolean } | null> => {
    try {
      const res = await fetch(`${API_URL}/debate/topics/${id}/conclude-vote`, {
        method: 'POST',
        headers: authHeaders(),
      })
      if (res.ok) {
        return await res.json()
      } else {
        const error = await res.json()
        throw new Error(error.error || 'Failed to vote to conclude')
      }
    } catch (e) {
      console.error('Failed to vote to conclude', e)
      return null
    }
  }

  const searchDebates = async (q: string, limit = 10): Promise<Debate[]> => {
    try {
      const res = await fetch(`${API_URL}/debate/topics/search?q=${encodeURIComponent(q)}&limit=${limit}`)
      if (res.ok) {
        const data = await res.json()
        return data.data || []
      }
      return []
    } catch (e) {
      console.error('Failed to search debates', e)
      return []
    }
  }

  const addDebateReference = async (argumentId: string, debateId: string): Promise<boolean> => {
    try {
      const res = await fetch(`${API_URL}/debate/arguments/${argumentId}/debate-reference`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', ...authHeaders() },
        body: JSON.stringify({ debate_id: debateId }),
      })
      return res.ok
    } catch (e) {
      console.error('Failed to add debate reference', e)
      return false
    }
  }

  const removeDebateReference = async (argumentId: string, debateId: string): Promise<boolean> => {
    try {
      const res = await fetch(
        `${API_URL}/debate/arguments/${argumentId}/debate-reference/${debateId}`,
        {
          method: 'DELETE',
          headers: authHeaders(),
        },
      )
      return res.ok
    } catch (e) {
      console.error('Failed to remove debate reference', e)
      return false
    }
  }

  const createArgument = async (
    debateId: string,
    payload: {
      content: string;
      argument_type:
        | "support"
        | "oppose"
        | "neutral"
        | "evidence"
        | "question"
        | "counter";
      parent_id?: string;
    },
  ): Promise<Argument | null> => {
    try {
      const res = await fetch(
        `${API_URL}/debate/topics/${debateId}/arguments`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json", ...authHeaders() },
          body: JSON.stringify(payload),
        },
      );
      if (!res.ok) {
        const error = await res.json();
        console.error("Create argument error:", error);
        throw new Error(error.error || "Failed to create argument");
      }
      const data = await res.json();
      return data.data as Argument;
    } catch (e) {
      console.error("Failed to create argument", e);
      throw e;
    }
  };

  const updateArgument = async (
    id: string,
    payload: {
      content: string;
      argument_type:
        | "support"
        | "oppose"
        | "neutral"
        | "evidence"
        | "question"
        | "counter";
      source_url?: string;
      source_title?: string;
      source_excerpt?: string;
    },
  ): Promise<Argument | null> => {
    try {
      const res = await fetch(`${API_URL}/debate/arguments/${id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json", ...authHeaders() },
        body: JSON.stringify(payload),
      });
      if (res.ok) {
        const data = await res.json();
        return data.data as Argument;
      } else {
        const error = await res.json();
        throw new Error(error.error || "Failed to update argument");
      }
    } catch (e) {
      console.error("Failed to update argument", e);
      return null;
    }
  };

  const deleteArgument = async (id: string): Promise<boolean> => {
    try {
      const res = await fetch(`${API_URL}/debate/arguments/${id}`, {
        method: "DELETE",
        headers: authHeaders(),
      });
      return res.ok;
    } catch (e) {
      console.error("Failed to delete argument", e);
      return false;
    }
  };

  const voteArgument = async (
    argumentId: string,
    voteType: number,
  ): Promise<Argument | null> => {
    try {
      const res = await fetch(
        `${API_URL}/debate/arguments/${argumentId}/vote`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json", ...authHeaders() },
          body: JSON.stringify({ vote_type: voteType }),
        },
      );
      if (res.ok) {
        const data = await res.json();
        // Track user vote locally
        userVotes.value[argumentId] = voteType;
        return data.data as Argument;
      } else {
        const error = await res.json();
        throw new Error(error.error || "Failed to vote");
      }
    } catch (e) {
      console.error("Failed to vote", e);
      return null;
    }
  };

  const removeVote = async (argumentId: string): Promise<boolean> => {
    try {
      const res = await fetch(
        `${API_URL}/debate/arguments/${argumentId}/vote`,
        {
          method: "DELETE",
          headers: authHeaders(),
        },
      );
      if (res.ok) {
        // Remove vote locally
        delete userVotes.value[argumentId];
      }
      return res.ok;
    } catch (e) {
      console.error("Failed to remove vote", e);
      return false;
    }
  };

  const addArgumentReference = async (
    argumentId: string,
    referenceId: string,
  ): Promise<boolean> => {
    try {
      const res = await fetch(
        `${API_URL}/debate/arguments/${argumentId}/reference`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json", ...authHeaders() },
          body: JSON.stringify({ reference_id: referenceId }),
        },
      );
      return res.ok;
    } catch (e) {
      console.error("Failed to add reference", e);
      return false;
    }
  };

  const removeArgumentReference = async (
    argumentId: string,
    referenceId: string,
  ): Promise<boolean> => {
    try {
      const res = await fetch(
        `${API_URL}/debate/arguments/${argumentId}/reference/${referenceId}`,
        {
          method: "DELETE",
          headers: authHeaders(),
        },
      );
      return res.ok;
    } catch (e) {
      console.error("Failed to remove reference", e);
      return false;
    }
  };

  const resetStore = () => {
    debates.value = [];
    debatesTotal.value = 0;
    currentDebate.value = null;
    argumentList.value = [];
    loading.value = false;
    error.value = null;
  };

  return {
    debates,
    debatesTotal,
    currentDebate,
    argumentList,
    loading,
    error,
    userVotes,
    fetchDebates,
    fetchDebate,
    fetchArguments,
    createDebate,
    updateDebate,
    deleteDebate,
    concludeDebate,
    reopenDebate,
    voteToConclude,
    searchDebates,
    createArgument,
    updateArgument,
    deleteArgument,
    voteArgument,
    removeVote,
    addArgumentReference,
    removeArgumentReference,
    addDebateReference,
    removeDebateReference,
    resetStore,
  }
})
